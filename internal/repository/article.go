package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"gorm.io/gorm"
	"log"
	"time"
	"vbook/internal/domain"
	"vbook/internal/repository/cache"
	"vbook/internal/repository/dao"
)

type ArticleRepository interface {
	Create(ctx context.Context, article domain.Article) (int64, error)
	Update(ctx context.Context, article domain.Article) error
	SyncV1(ctx context.Context, article domain.Article) (int64, error)
	Sync(ctx context.Context, article domain.Article) (int64, error)
	SyncStatus(ctx context.Context, uid int64, id int64, private domain.ArticleStatus) error
	GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error)
	GetById(ctx context.Context, id int64) (domain.Article, error)
	GetPubById(ctx context.Context, id int64) (domain.Article, error)
}
type CacheArticleRepository struct {
	ad        dao.ArticleDao
	readerDao dao.ArticleReaderDao
	authorDao dao.ArticleAuthorDao
	db        *gorm.DB
	cache     cache.ArticleCache
	userRepo  UserRepository
}

func NewArticleRepositoryV2(readerDao dao.ArticleReaderDao, authorDao dao.ArticleAuthorDao) *CacheArticleRepository {
	return &CacheArticleRepository{
		readerDao: readerDao,
		authorDao: authorDao,
	}
}
func NewArticleRepository(ad dao.ArticleDao, cache cache.ArticleCache) ArticleRepository {
	return &CacheArticleRepository{
		ad:    ad,
		cache: cache,
	}
}
func (ar *CacheArticleRepository) Create(ctx context.Context, article domain.Article) (int64, error) {
	id, err := ar.ad.Insert(ctx, ar.toDao(article))
	if err == nil {
		err := ar.cache.DelFirstPage(ctx, article.Author.Id)
		if err != nil {
			log.Println()
		}
	}
	return id, err
}
func (ar *CacheArticleRepository) toDao(article domain.Article) dao.Article {
	return dao.Article{
		Id:       article.Id,
		Title:    article.Title,
		Content:  article.Content,
		AuthorId: article.Author.Id,
		Status:   uint8(article.Status),
	}
}
func (ar *CacheArticleRepository) toDomain(art dao.Article) domain.Article {
	return domain.Article{
		Id:      art.Id,
		Title:   art.Title,
		Content: art.Content,
		Author: domain.Author{
			Id: art.AuthorId,
		},
		Ctime:  time.UnixMilli(art.Ctime),
		Utime:  time.UnixMilli(art.Utime),
		Status: domain.ArticleStatus(art.Status),
	}
}
func (ar *CacheArticleRepository) Update(ctx context.Context, article domain.Article) error {
	err := ar.ad.UpdateById(ctx, ar.toDao(article))
	if err == nil {
		err := ar.cache.DelFirstPage(ctx, article.Author.Id)
		if err != nil {
			log.Println()
		}
	}
	return err
}
func (ar *CacheArticleRepository) Sync(ctx context.Context, article domain.Article) (int64, error) {
	id, err := ar.ad.Sync(ctx, ar.toDao(article))
	if err == nil {
		err := ar.cache.DelFirstPage(ctx, article.Author.Id)
		if err != nil {
			log.Println()
		}
	}
	//在这里设置缓存
	go func() {
		user, err := ar.userRepo.FindById(ctx, article.Author.Id)
		if err != nil {
			log.Println(err)
			return
		}
		article.Author = domain.Author{
			Id:   user.Id,
			Name: user.Name,
		}
		err = ar.cache.SetPub(ctx, article)
		if err != nil {
			log.Println(err)
		}
	}()
	return id, err
}
func (ar *CacheArticleRepository) SyncV1(ctx context.Context, article domain.Article) (int64, error) {
	artDao := ar.toDao(article)
	var (
		id  = artDao.Id
		err error
	)
	if id > 0 {
		err = ar.authorDao.Update(ctx, artDao)
	} else {
		id, err = ar.authorDao.Create(ctx, artDao)
	}
	if err != nil {
		return 0, nil
	}
	artDao.Id = id
	err = ar.readerDao.Save(ctx, artDao)
	return id, err
}
func (ar *CacheArticleRepository) SyncV2(ctx context.Context, article domain.Article) (int64, error) {
	tx := ar.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, tx.Error
	}
	//防止后面业务panic
	defer tx.Rollback()
	authorDao := dao.NewGormArticleAuthorDao(tx)
	readerDao := dao.NewGormArticleReaderDao(tx)
	artDao := ar.toDao(article)
	var (
		id  = artDao.Id
		err error
	)
	if id > 0 {
		err = authorDao.Update(ctx, artDao)
	} else {
		id, err = authorDao.Create(ctx, artDao)
	}
	if err != nil {
		return 0, err
	}
	artDao.Id = id
	err = readerDao.SaveV2(ctx, dao.PublishedArticle(artDao))
	if err != nil {
		return 0, err
	}
	tx.Commit()
	return id, nil
}
func (ar *CacheArticleRepository) SyncStatus(ctx context.Context, uid int64, id int64, private domain.ArticleStatus) error {
	err := ar.ad.SyncStatus(ctx, uid, id, uint8(private))
	if err == nil {
		err := ar.cache.DelFirstPage(ctx, uid)
		if err != nil {
			log.Println()
		}
	}
	return err
}
func (ar *CacheArticleRepository) GetByAuthor(ctx context.Context, uid int64, offset int, limit int) ([]domain.Article, error) {
	//第一步，判定要不要缓存
	if offset == 0 && limit == 100 {
		res, err := ar.cache.GetFirstPage(ctx, uid)
		if err == nil {
			return res, err
		} else {
			log.Println()
		}
	}
	arts, err := ar.ad.GetByAuthor(ctx, uid, offset, limit)
	if err != nil {
		return nil, err
	}
	res := slice.Map[dao.Article, domain.Article](arts, func(idx int, src dao.Article) domain.Article {
		return ar.toDomain(src)
	})
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if offset == 0 && limit == 100 {
			err = ar.cache.SetFirstPage(ctx, uid, res)
			if err != nil {
				log.Println(err)
				//监控这里
			}
		}
	}()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ar.preCache(ctx, res)
	}()
	return res, nil
}
func (ar *CacheArticleRepository) GetById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := ar.cache.Get(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := ar.ad.GetById(ctx, id)
	res = ar.toDomain(art)
	if err != nil {
		return domain.Article{}, err
	}
	go func() {
		err := ar.cache.Set(ctx, res)
		if err != nil {
			log.Println(err)
		}
	}()
	return res, nil
}
func (ar *CacheArticleRepository) preCache(ctx context.Context, arts []domain.Article) {
	const size = 1024 * 1024
	if len(arts) > 0 && len(arts[0].Content) <= size {
		err := ar.cache.Set(ctx, arts[0])
		if err != nil {
			log.Println(err)
		}
	}
}
func (ar *CacheArticleRepository) GetPubById(ctx context.Context, id int64) (domain.Article, error) {
	res, err := ar.cache.GetPub(ctx, id)
	if err == nil {
		return res, nil
	}
	art, err := ar.ad.GetPubById(ctx, id)
	if err != nil {
		return domain.Article{}, err
	}
	res = ar.toDomain(dao.Article(art))
	author, err := ar.userRepo.FindById(ctx, art.AuthorId)
	if err != nil {
		log.Println(err)
		return domain.Article{}, err
	}
	res.Author.Name = author.Name
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		ar.cache.SetPub(ctx, res)
		if err != nil {
			log.Println(err)
		}
	}()
	return res, nil
}

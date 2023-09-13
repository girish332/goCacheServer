package cache

import (
	"cacheServer/appcontext"
	"cacheServer/apperror"
	"database/sql"
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"log"
	"regexp"
	"testing"
	"time"
)

func newMock() *dbMock {
	db, mock, err := sqlmock.New()
	if err != nil {
		log.Println("an error occured while creating new mocksql")
	}
	m := newDbMock(db, mock)
	return m
}

type dbMock struct {
	db      *sql.DB
	mocksql sqlmock.Sqlmock
}

func newDbMock(db *sql.DB, mock sqlmock.Sqlmock) *dbMock {
	return &dbMock{db: db, mocksql: mock}
}

var (
	db                         = newMock()
	appCtx *appcontext.Context = appcontext.NewContext(db.db, 1)
	s      *Server             = GetCacheInstance(appCtx)
)

func setUp() {
	go s.Run()
}

func tearDown() {
	s.Close()
}

func TestRunProduct(t *testing.T) {
	cases := map[string]struct {
		want       bool
		err        error
		testTarget string
		request    *Request
	}{
		"when productID present in DB but not in cache": {
			want:       true,
			err:        nil,
			testTarget: "db",
			request:    NewRequest("test1", Product, nil),
		},
		"when productID not present in DB": {
			want:       false,
			err:        errors.New("error"),
			testTarget: "db",
			request:    NewRequest("test2", Product, nil),
		},
		"when productID is present in cache": {
			want:       true,
			err:        nil,
			testTarget: "cache",
			request:    NewRequest("test3", Product, nil),
		},
	}

	setUp()
	defer tearDown()
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			switch v.testTarget {
			case "db":
				if v.want {
					query := `SELECT id FROM "products" WHERE id=$1;`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("test1"))
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				} else if !v.want {
					query := `SELECT id FROM "products" WHERE id=$1;`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test2").WillReturnError(v.err)
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			case "cache":
				if v.want {
					s.store.data[Product]["test3"] = "active"
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			}
		})
	}
}

func TestRunCacheCategory(t *testing.T) {
	cases := map[string]struct {
		want       bool
		err        error
		testTarget string
		request    *Request
		t          Type
	}{
		"when categoryID present in DB but not in cache": {
			want:       true,
			err:        nil,
			testTarget: "db",
			request:    NewRequest("test1", Category, nil),
		},
		"when categoryID not present in DB": {
			want:       false,
			err:        errors.New("error"),
			testTarget: "db",
			request:    NewRequest("test2", Category, nil),
		},
		"when categoryID is present in cache": {
			want:       true,
			err:        nil,
			testTarget: "cache",
			request:    NewRequest("test3", Category, nil),
		},
	}
	setUp()
	defer tearDown()
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			switch v.testTarget {
			case "db":
				if v.want {
					query := `SELECT id FROM "productCategory" WHERE id=$1;`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("test1"))
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				} else if !v.want {
					query := `SELECT id FROM "productCategory" WHERE id=$1;`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test2").WillReturnError(v.err)
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			case "cache":
				if v.want {
					s.store.data[Category]["test3"] = "active"
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			}
		})
	}
}

func TestUpdateCache(t *testing.T) {
	setUp()
	defer tearDown()
	s.updateCache("test", "test", Role)
}

func TestDeleteCache(t *testing.T) {
	setUp()
	defer tearDown()
	s.DeleteCache("test", Category)
}

func TestRunSubcategory(t *testing.T) {
	cases := map[string]struct {
		want       bool
		err        error
		testTarget string
		request    *Request
	}{
		"when subcategoryID present in DB but not in cache": {
			want:       true,
			err:        nil,
			testTarget: "db",
			request:    NewRequest("test1", Subcategory, nil),
		},
		"when subcategoryID not present in DB": {
			want:       false,
			err:        errors.New("error"),
			testTarget: "db",
			request:    NewRequest("test2", Subcategory, nil),
		},
		"when subcategoryID is present in cache": {
			want:       true,
			err:        nil,
			testTarget: "cache",
			request:    NewRequest("test3", Subcategory, nil),
		},
	}
	setUp()
	defer tearDown()
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			switch v.testTarget {
			case "db":
				if v.want {
					query := `SELECT id FROM "productSubCategory" WHERE id=$1;`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test1").WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow("test1"))
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				} else if !v.want {
					query := `SELECT id FROM "productSubCategory" WHERE id=$1;`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test2").WillReturnError(v.err)
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			case "cache":
				if v.want {
					s.store.data[Subcategory]["test3"] = "active"
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			}
		})
	}
}

func TestRunRole(t *testing.T) {
	cases := map[string]struct {
		want       bool
		err        error
		request    *Request
		testTarget string
	}{
		"when correct role is passed and is not present in cache": {
			want:       true,
			err:        nil,
			request:    NewRequest("test1", Role, "test1"),
			testTarget: "db",
		},
		"when database throws error": {
			want:       false,
			err:        errors.New("error"),
			request:    NewRequest("test123", Role, "test"),
			testTarget: "db",
		},
		"if role present in cache": {
			want:       true,
			err:        nil,
			request:    NewRequest("test3", Role, "admin"),
			testTarget: "cache",
		},
		"when opt is passed as nil": {
			want:       false,
			err:        errors.New("error"),
			request:    NewRequest("test3", Role, nil),
			testTarget: "cache",
		},
	}
	setUp()
	defer tearDown()
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			switch v.testTarget {
			case "db":
				if v.want {
					query := `SELECT "role" FROM "users" WHERE "emailId" = $1`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test1").WillReturnRows(sqlmock.NewRows([]string{"test1"}).AddRow("test1"))
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, true, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				} else if !v.want {
					query := `SELECT "role" FROM "users" WHERE "emailId" = $1`
					prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
					prep.WithArgs("test123").WillReturnError(v.err)
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			case "cache":
				if v.want {
					s.store.data[Role]["test3"] = "admin"
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				} else if !v.want {
					s.store.data[Role]["test4"] = "admin"
					s.MakeRequest(v.request)
					time.Sleep(1 * time.Millisecond)
					assert.Equal(t, v.want, <-v.request.Out)
					time.Sleep(1 * time.Millisecond)
				}
			}
		})
	}
}

func TestInitializeSubcategoryCache(t *testing.T) {
	query := `SELECT "categoryID",ARRAY_AGG("index") FROM (
            SELECT "categoryID","index" FROM "productSubCategory" GROUP BY 1,2 ORDER BY 2 ASC) t1
            GROUP BY 1;`

	testCache := make(map[string][255]bool)
	var result [255]bool
	result[3] = true
	result[4] = true
	result[7] = true
	testCache["test"] = result
	cases := map[string]struct {
		want     map[string][255]bool
		getErr   error
		prepFunc queryPrepareFunc
	}{
		"success": {
			want:   testCache,
			getErr: nil,
			prepFunc: func(prep *sqlmock.ExpectedQuery) {
				prep.WillReturnRows(sqlmock.NewRows([]string{"categoryID", "index"}).AddRow("test", (pq.Int32Array)([]int32{1, 2, 5, 6})))
			},
		},
	}
	setUp()
	defer tearDown()

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			time.Sleep(1 * time.Millisecond)
			prep1 := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
			v.prepFunc(prep1)
			s.initializeSubcategoryCache()
			assert.Equal(t, v.want, s.store.subcategoryIndices)
		})
	}
	delete(s.store.subcategoryIndices, "test")
}

type queryPrepareFunc = func(prep *sqlmock.ExpectedQuery)

func TestInitializeProductCache(t *testing.T) {
	testMap := make(map[int]struct{})
	testMap[2] = struct{}{}
	p := NewSortedIndices(make([]int, 0))
	s.store.productIndices["test4"] = p

	cases := map[string]struct {
		want      map[int]struct{}
		getErr    error
		prepFunc1 queryPrepareFunc
		prepFunc2 queryPrepareFunc
	}{
		"success": {
			want:   testMap,
			getErr: nil,
			prepFunc1: func(prep *sqlmock.ExpectedQuery) {
				prep.WillReturnRows(sqlmock.NewRows([]string{"subcategoryID", "index"}).AddRow("test4", (pq.Int32Array)([]int32{1})))
			},
			prepFunc2: func(prep *sqlmock.ExpectedQuery) {
				prep.WillReturnRows(sqlmock.NewRows([]string{"subcategoryID"}).AddRow("test"))
			},
		},
	}
	query := `SELECT "subCategoryID",ARRAY_AGG("index") FROM (
              SELECT "subCategoryID","index" FROM "products" GROUP BY 1,2 ORDER BY 2 ASC) t1 
              GROUP BY 1;`
	query2 := `SELECT id FROM "productSubCategory";`
	setUp()
	defer tearDown()

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			prep2 := db.mocksql.ExpectQuery(regexp.QuoteMeta(query2))
			v.prepFunc2(prep2)
			prep1 := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
			v.prepFunc1(prep1)
			s.initializeProductCache()
			assert.Equal(t, v.want, s.store.productIndices["test4"].availableIndices)
		})
	}
	time.Sleep(10 * time.Millisecond)
	delete(s.store.productIndices, "test4")
}

func TestDeleteCategoryIndexCache(t *testing.T) {
	var result [255]bool
	result[1] = true
	query := `SELECT index from "productCategory" ORDER BY index ASC;`
	cases := map[string]struct {
		want           bool
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: false,
			err:  nil,
			initialization: func() {
				s.store.categoryIndices = result
			},
		},
		"cache is not initialized": {
			want: false,
			err:  apperror.ErrCacheNotInitialized,
			initialization: func() {
				s.store.categoryIndices = [255]bool{}
				prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
				prep.WillReturnError(apperror.ErrCacheNotInitialized)
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			err := s.DeleteCategoryIndexCache(1)
			assert.Equal(t, v.err, err)
			assert.Equal(t, v.want, s.store.categoryIndices[1])
			s.store.categoryIndices = [255]bool{}
		})
	}
}

func TestUpdateCategoryIndexCache(t *testing.T) {
	cases := map[string]struct {
		want           bool
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: true,
			err:  nil,
			initialization: func() {
				s.store.categoryIndices[1] = true
			},
		},
		"cache is not initialized": {
			want:           false,
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			err := s.UpdateCategoryIndexCache(3)
			assert.Equal(t, err, v.err)
			assert.Equal(t, s.store.categoryIndices[3], v.want)
			s.store.categoryIndices = [255]bool{}
		})
	}
}

func TestGetCategoryIndicesCache(t *testing.T) {
	response := []int{1}
	cases := map[string]struct {
		want           []int
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: response,
			err:  nil,
			initialization: func() {
				s.store.categoryIndices = [255]bool{}
				s.store.categoryIndices[1] = true
			},
		},
		"cache is not initialized": {
			want:           nil,
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			res, err := s.GetCategoryIndicesCache()
			assert.Equal(t, err, v.err)
			assert.Equal(t, res, v.want)
			s.store.categoryIndices = [255]bool{}
		})
	}
}

func TestGetMaximumIndexCategory(t *testing.T) {
	cases := map[string]struct {
		want           int
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: 1,
			err:  nil,
			initialization: func() {
				s.store.categoryIndices = [255]bool{}
				s.store.categoryIndices[1] = true
			},
		},
		"cache is not initialized": {
			want:           0,
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			get, err := s.GetMaximumIndexCategory()
			assert.Equal(t, err, v.err)
			assert.Equal(t, get, v.want)
			s.store.categoryIndices = [255]bool{}
		})
	}
}

func TestInitializeCategoryCache(t *testing.T) {
	query := `SELECT index from "productCategory" ORDER BY index ASC;`
	prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
	prep.WillReturnRows(sqlmock.NewRows([]string{"index"}).AddRow(1).AddRow(3))
	s.store.categoryIndices = [255]bool{}

	arr := [255]bool{}
	arr[2] = true
	arr[4] = true
	cases := map[string]struct {
		want     [255]bool
		getErr   error
		prepFunc queryPrepareFunc
	}{
		"success": {
			want:   arr,
			getErr: nil,
			prepFunc: func(prep *sqlmock.ExpectedQuery) {
				prep.WillReturnRows(sqlmock.NewRows([]string{"index"}).AddRow(1).AddRow(3))
			},
		},
	}
	setUp()
	defer tearDown()

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			prep := db.mocksql.ExpectQuery(regexp.QuoteMeta(query))
			v.prepFunc(prep)
			s.initializeCategoryCache()
			assert.Equal(t, v.want, s.store.categoryIndices)
		})
	}
	s.store.categoryIndices = [255]bool{}
}

func TestCreateSubcategoryCache(t *testing.T) {
	var result [255]bool
	result[1] = true
	cases := map[string]struct {
		want [255]bool
		err  error
	}{
		"cache is initialized": {
			want: result,
			err:  nil,
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			err := s.CreateSubcategoryCache("test3")
			assert.Equal(t, err, v.err)
			assert.Equal(t, s.store.subcategoryIndices["test3"], v.want)
			delete(s.store.subcategoryIndices, "test3")
		})
	}
}

func TestUpdateSubcategoryIndexCache(t *testing.T) {
	var result [255]bool
	result[1] = true
	cases := map[string]struct {
		want           [255]bool
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: result,
			err:  nil,
			initialization: func() {
				s.store.subcategoryIndices["test"] = result
			},
		},
		"cache is not initialized": {
			want:           [255]bool{},
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			err := s.UpdateSubcategoryIndexCache(1, "test")
			assert.Equal(t, err, v.err)
			assert.Equal(t, s.store.subcategoryIndices["test"], v.want)
			delete(s.store.subcategoryIndices, "test")
		})
	}
}

func TestDeleteSubcategoryIndexCache(t *testing.T) {
	var result [255]bool
	result[1] = true
	cases := map[string]struct {
		want           [255]bool
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: [255]bool{},
			err:  nil,
			initialization: func() {
				s.store.subcategoryIndices["test"] = result
			},
		},
		"cache is not initialized": {
			want:           [255]bool{},
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			err := s.DeleteSubcategoryIndexCache("test", 1)
			assert.Equal(t, err, v.err)
			assert.Equal(t, s.store.subcategoryIndices["test"], v.want)
			delete(s.store.subcategoryIndices, "test")
		})
	}
}

func TestGetSubcategoryIndicesCache(t *testing.T) {
	var result [255]bool
	result[1] = true
	response := []int{1}
	cases := map[string]struct {
		want           []int
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: response,
			err:  nil,
			initialization: func() {
				s.store.subcategoryIndices["test"] = result
			},
		},
		"cache is not initialized": {
			want:           nil,
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}
	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			get, err := s.GetSubcategoryIndicesCache("test")
			assert.Equal(t, err, v.err)
			assert.Equal(t, get, v.want)
			delete(s.store.subcategoryIndices, "test")
		})
	}
}

func TestGetMaximumIndexSubcategory(t *testing.T) {
	var result [255]bool
	result[1] = true
	cases := map[string]struct {
		want           int
		err            error
		initialization func()
	}{
		"cache is initialized": {
			want: 1,
			err:  nil,
			initialization: func() {
				s.store.subcategoryIndices["test"] = result
			},
		},
		"cache is not initialized": {
			want:           0,
			err:            apperror.ErrCacheNotInitialized,
			initialization: func() {},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			get, err := s.GetMaximumIndexSubcategory("test")
			assert.Equal(t, err, v.err)
			assert.Equal(t, get, v.want)
			delete(s.store.subcategoryIndices, "test")
		})
	}
}

func TestCreateProductCache(t *testing.T) {
	s.CreateProductCache("test5")
	assert.Equal(t, s.store.productIndices["test5"].orderedIndices, []int{1})
	delete(s.store.subcategoryIndices, "test5")
}

func TestUpdateProductCacheIndex(t *testing.T) {
	testMap := make(map[int]struct{})
	testMap[1] = struct{}{}
	testMap[2] = struct{}{}

	wrongMap := make(map[int]struct{})
	cases := map[string]struct {
		want           map[int]struct{}
		err            error
		initialization func()
	}{
		"success": {
			want: testMap,
			err:  nil,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test4"] = p
				s.store.productIndices["test4"].availableIndices[2] = struct{}{}
				s.store.productIndices["test4"].orderedIndices = []int{2}
			},
		},
		"cache not initialized": {
			want: wrongMap,
			err:  apperror.ErrCacheNotInitialized,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test4"] = p
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			err := s.UpdateProductCacheIndex(1, "test4")
			time.Sleep(10 * time.Millisecond)
			assert.Equal(t, v.err, err)
			assert.Equal(t, v.want, s.store.productIndices["test4"].availableIndices)
			delete(s.store.productIndices, "test4")
		})
	}
}

func TestDeleteProductCacheIndex(t *testing.T) {
	testMap := make(map[int]struct{})
	testMap[1] = struct{}{}
	wrongMap := make(map[int]struct{})

	cases := map[string]struct {
		want           map[int]struct{}
		err            error
		initialization func()
	}{
		"success": {
			want: testMap,
			err:  nil,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test3"] = p
				s.store.productIndices["test3"].availableIndices[1] = struct{}{}
				s.store.productIndices["test3"].availableIndices[2] = struct{}{}
				s.store.productIndices["test3"].orderedIndices = []int{1, 2}
			},
		},
		"cache not initialized": {
			want: wrongMap,
			err:  apperror.ErrCacheNotInitialized,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test3"] = p
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			err := s.DeleteProductCacheIndex("test3", 2)
			assert.Equal(t, v.err, err)
			assert.Equal(t, v.want, s.store.productIndices["test3"].availableIndices)
			time.Sleep(10 * time.Millisecond)
			delete(s.store.productIndices, "test3")
		})
	}

}

func TestGetProductIndicesCache(t *testing.T) {

	cases := map[string]struct {
		want           []int
		err            error
		initialization func()
	}{
		"success": {
			want: []int{1},
			err:  nil,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test2"] = p
				s.store.productIndices["test2"].availableIndices[1] = struct{}{}
				s.store.productIndices["test2"].orderedIndices = []int{1}
			},
		},
		"cache not initialized": {
			want: nil,
			err:  apperror.ErrCacheNotInitialized,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test2"] = p
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			get, err := s.GetProductIndicesCache("test2")
			assert.Equal(t, v.err, err)
			assert.Equal(t, v.want, get)
			time.Sleep(10 * time.Millisecond)
			delete(s.store.productIndices, "test2")
		})
	}

}

func TestGetMaximumIndexProduct(t *testing.T) {
	cases := map[string]struct {
		want           int
		err            error
		initialization func()
	}{
		"success": {
			want: 1,
			err:  nil,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test1"] = p
				s.store.productIndices["test1"].availableIndices[1] = struct{}{}
				s.store.productIndices["test1"].orderedIndices = []int{1}
			},
		},
		"cache not initialized": {
			want: 0,
			err:  apperror.ErrCacheNotInitialized,
			initialization: func() {
				p := NewSortedIndices(make([]int, 0))
				s.store.productIndices["test1"] = p
			},
		},
	}

	for k, v := range cases {
		t.Run(k, func(t *testing.T) {
			v.initialization()
			get, err := s.GetMaximumIndexProduct("test1")
			assert.Equal(t, v.err, err)
			assert.Equal(t, v.want, get)
			time.Sleep(10 * time.Millisecond)
			delete(s.store.productIndices, "test1")
		})
	}

}

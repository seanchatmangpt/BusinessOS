package utils

import (
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gin-gonic/gin"
)

// setupPaginationContext creates a gin context with the given query parameters.
func setupPaginationContext(query string) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items?"+query, nil)
	return c, w
}

// ---------------------------------------------------------------------------
// DefaultPaginationParams
// ---------------------------------------------------------------------------

func TestDefaultPaginationParams_ReturnsPage1PerPage20(t *testing.T) {
	p := DefaultPaginationParams()

	if p.Page != 1 {
		t.Errorf("expected Page=1, got %d", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("expected PerPage=20, got %d", p.PerPage)
	}
}

// ---------------------------------------------------------------------------
// ExtractPaginationParams — valid inputs
// ---------------------------------------------------------------------------

func TestExtractPaginationParams_ValidPageAndPerPage(t *testing.T) {
	c, _ := setupPaginationContext("page=3&per_page=50")

	p := ExtractPaginationParams(c)

	if p.Page != 3 {
		t.Errorf("expected Page=3, got %d", p.Page)
	}
	if p.PerPage != 50 {
		t.Errorf("expected PerPage=50, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_NoQueryParams_ReturnsDefaults(t *testing.T) {
	c, _ := setupPaginationContext("")

	p := ExtractPaginationParams(c)

	if p.Page != 1 {
		t.Errorf("expected default Page=1, got %d", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_OnlyPageProvided(t *testing.T) {
	c, _ := setupPaginationContext("page=5")

	p := ExtractPaginationParams(c)

	if p.Page != 5 {
		t.Errorf("expected Page=5, got %d", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_OnlyPerPageProvided(t *testing.T) {
	c, _ := setupPaginationContext("per_page=10")

	p := ExtractPaginationParams(c)

	if p.Page != 1 {
		t.Errorf("expected default Page=1, got %d", p.Page)
	}
	if p.PerPage != 10 {
		t.Errorf("expected PerPage=10, got %d", p.PerPage)
	}
}

// ---------------------------------------------------------------------------
// ExtractPaginationParams — edge cases
// ---------------------------------------------------------------------------

func TestExtractPaginationParams_NegativePage_IgnoresAndReturnsDefault(t *testing.T) {
	c, _ := setupPaginationContext("page=-1&per_page=20")

	p := ExtractPaginationParams(c)

	if p.Page != 1 {
		t.Errorf("expected default Page=1 for negative input, got %d", p.Page)
	}
}

func TestExtractPaginationParams_ZeroPage_IgnoresAndReturnsDefault(t *testing.T) {
	c, _ := setupPaginationContext("page=0")

	p := ExtractPaginationParams(c)

	if p.Page != 1 {
		t.Errorf("expected default Page=1 for zero input, got %d", p.Page)
	}
}

func TestExtractPaginationParams_NegativePerPage_IgnoresAndReturnsDefault(t *testing.T) {
	c, _ := setupPaginationContext("per_page=-5")

	p := ExtractPaginationParams(c)

	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20 for negative input, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_ZeroPerPage_IgnoresAndReturnsDefault(t *testing.T) {
	c, _ := setupPaginationContext("per_page=0")

	p := ExtractPaginationParams(c)

	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20 for zero input, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_NonNumericPage_ReturnsDefault(t *testing.T) {
	c, _ := setupPaginationContext("page=abc")

	p := ExtractPaginationParams(c)

	if p.Page != 1 {
		t.Errorf("expected default Page=1 for non-numeric input, got %d", p.Page)
	}
}

func TestExtractPaginationParams_NonNumericPerPage_ReturnsDefault(t *testing.T) {
	c, _ := setupPaginationContext("per_page=xyz")

	p := ExtractPaginationParams(c)

	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20 for non-numeric input, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_PerPageExceedsMax_CappedAt100(t *testing.T) {
	c, _ := setupPaginationContext("per_page=500")

	p := ExtractPaginationParams(c)

	if p.PerPage != 100 {
		t.Errorf("expected PerPage capped at 100, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_PerPageAtMaxBoundary(t *testing.T) {
	c, _ := setupPaginationContext("per_page=100")

	p := ExtractPaginationParams(c)

	if p.PerPage != 100 {
		t.Errorf("expected PerPage=100 at boundary, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_PerPageJustAboveMax(t *testing.T) {
	c, _ := setupPaginationContext("per_page=101")

	p := ExtractPaginationParams(c)

	if p.PerPage != 100 {
		t.Errorf("expected PerPage=100 for 101 input, got %d", p.PerPage)
	}
}

func TestExtractPaginationParams_LargePageNumber(t *testing.T) {
	c, _ := setupPaginationContext("page=999999")

	p := ExtractPaginationParams(c)

	if p.Page != 999999 {
		t.Errorf("expected Page=999999, got %d", p.Page)
	}
}

func TestExtractPaginationParams_FloatPage_Truncates(t *testing.T) {
	c, _ := setupPaginationContext("page=3.7")

	p := ExtractPaginationParams(c)

	// strconv.ParseInt with base 10 will fail on "3.7", so default is used
	if p.Page != 1 {
		t.Errorf("expected default Page=1 for float input, got %d", p.Page)
	}
}

func TestExtractPaginationParams_EmptyStringValues_ReturnsDefaults(t *testing.T) {
	c, _ := setupPaginationContext("page=&per_page=")

	p := ExtractPaginationParams(c)

	if p.Page != 1 {
		t.Errorf("expected default Page=1 for empty string, got %d", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("expected default PerPage=20 for empty string, got %d", p.PerPage)
	}
}

// ---------------------------------------------------------------------------
// Offset / Limit
// ---------------------------------------------------------------------------

func TestOffset_FirstPage_ReturnsZero(t *testing.T) {
	p := PaginationParams{Page: 1, PerPage: 20}
	if p.Offset() != 0 {
		t.Errorf("expected offset 0 for page 1, got %d", p.Offset())
	}
}

func TestOffset_SecondPage_ReturnsPerPage(t *testing.T) {
	p := PaginationParams{Page: 2, PerPage: 20}
	if p.Offset() != 20 {
		t.Errorf("expected offset 20 for page 2, got %d", p.Offset())
	}
}

func TestOffset_TenthPage_ReturnsCorrectOffset(t *testing.T) {
	p := PaginationParams{Page: 10, PerPage: 25}
	expected := int32(9 * 25) // 225
	if p.Offset() != expected {
		t.Errorf("expected offset %d, got %d", expected, p.Offset())
	}
}

func TestOffset_LargePageNumbers(t *testing.T) {
	p := PaginationParams{Page: math.MaxInt32, PerPage: math.MaxInt32}
	// (MaxInt32 - 1) * MaxInt32 will overflow int32, but Go handles it as
	// two's complement wrap. The important thing is it doesn't panic.
	_ = p.Offset() // should not panic
}

func TestLimit_ReturnsPerPage(t *testing.T) {
	p := PaginationParams{Page: 3, PerPage: 50}
	if p.Limit() != 50 {
		t.Errorf("expected limit=50, got %d", p.Limit())
	}
}

func TestLimit_OnePerPage(t *testing.T) {
	p := PaginationParams{Page: 1, PerPage: 1}
	if p.Limit() != 1 {
		t.Errorf("expected limit=1, got %d", p.Limit())
	}
}

// ---------------------------------------------------------------------------
// SetPaginationHeaders
// ---------------------------------------------------------------------------

func TestSetPaginationHeaders_SetsCorrectHeaders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items", nil)

	params := PaginationParams{Page: 2, PerPage: 50}
	SetPaginationHeaders(c, 1000, params)

	if w.Header().Get("X-Total-Count") != "1000" {
		t.Errorf("expected X-Total-Count=1000, got %s", w.Header().Get("X-Total-Count"))
	}
	if w.Header().Get("X-Page") != "2" {
		t.Errorf("expected X-Page=2, got %s", w.Header().Get("X-Page"))
	}
	if w.Header().Get("X-Per-Page") != "50" {
		t.Errorf("expected X-Per-Page=50, got %s", w.Header().Get("X-Per-Page"))
	}
}

func TestSetPaginationHeaders_ZeroTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items", nil)

	SetPaginationHeaders(c, 0, DefaultPaginationParams())

	if w.Header().Get("X-Total-Count") != "0" {
		t.Errorf("expected X-Total-Count=0, got %s", w.Header().Get("X-Total-Count"))
	}
}

func TestSetPaginationHeaders_LargeTotal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/items", nil)

	largeTotal := int64(math.MaxInt64)
	SetPaginationHeaders(c, largeTotal, DefaultPaginationParams())

	expected := strconv.FormatInt(largeTotal, 10)
	if w.Header().Get("X-Total-Count") != expected {
		t.Errorf("expected X-Total-Count=%s, got %s", expected, w.Header().Get("X-Total-Count"))
	}
}

// ---------------------------------------------------------------------------
// NewPaginatedResponse
// ---------------------------------------------------------------------------

func TestNewPaginatedResponse_ReturnsCorrectStructure(t *testing.T) {
	items := []string{"a", "b", "c"}
	params := PaginationParams{Page: 1, PerPage: 20}

	resp := NewPaginatedResponse(items, int64(len(items)), params)

	if resp.Total != 3 {
		t.Errorf("expected total=3, got %d", resp.Total)
	}
	if resp.Page != 1 {
		t.Errorf("expected page=1, got %d", resp.Page)
	}
	if resp.PerPage != 20 {
		t.Errorf("expected per_page=20, got %d", resp.PerPage)
	}
	if resp.Items == nil {
		t.Error("expected items to be non-nil")
	}
}

func TestNewPaginatedResponse_WithEmptyItems(t *testing.T) {
	resp := NewPaginatedResponse([]int{}, 0, DefaultPaginationParams())

	if resp.Total != 0 {
		t.Errorf("expected total=0, got %d", resp.Total)
	}
	if resp.Items == nil {
		t.Error("expected items to be non-nil (empty slice)")
	}
}

func TestNewPaginatedResponse_WithNilItems(t *testing.T) {
	resp := NewPaginatedResponse(nil, 0, DefaultPaginationParams())

	if resp.Total != 0 {
		t.Errorf("expected total=0, got %d", resp.Total)
	}
	if resp.Items != nil {
		t.Error("expected items to be nil when passed nil")
	}
}

func TestNewPaginatedResponse_Page2(t *testing.T) {
	resp := NewPaginatedResponse([]int{21, 22, 23}, 100, PaginationParams{Page: 2, PerPage: 3})

	if resp.Page != 2 {
		t.Errorf("expected page=2, got %d", resp.Page)
	}
	if resp.PerPage != 3 {
		t.Errorf("expected per_page=3, got %d", resp.PerPage)
	}
	if resp.Total != 100 {
		t.Errorf("expected total=100, got %d", resp.Total)
	}
}

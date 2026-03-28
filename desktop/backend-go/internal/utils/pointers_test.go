package utils

import (
	"math"
	"testing"
)

// ---------------------------------------------------------------------------
// StringPtr
// ---------------------------------------------------------------------------

func TestStringPtr_ReturnsNonNullPointer(t *testing.T) {
	p := StringPtr("hello")

	if p == nil {
		t.Fatal("expected non-nil pointer, got nil")
	}
	if *p != "hello" {
		t.Errorf("expected %q, got %q", "hello", *p)
	}
}

func TestStringPtr_EmptyString(t *testing.T) {
	p := StringPtr("")

	if p == nil {
		t.Fatal("expected non-nil pointer for empty string, got nil")
	}
	if *p != "" {
		t.Errorf("expected empty string, got %q", *p)
	}
}

func TestStringPtr_DoesNotShareMemory(t *testing.T) {
	p1 := StringPtr("original")
	p2 := StringPtr("original")

	// Modifying via one pointer must not affect the other.
	*p1 = "changed"

	if *p2 == "changed" {
		t.Error("expected independent pointers; modifying p1 affected p2")
	}
}

func TestStringPtr_UnicodeContent(t *testing.T) {
	p := StringPtr("Hello, world!")

	if *p != "Hello, world!" {
		t.Errorf("expected %q, got %q", "Hello, world!", *p)
	}
}

// ---------------------------------------------------------------------------
// IntPtr
// ---------------------------------------------------------------------------

func TestIntPtr_ReturnsNonNullPointer(t *testing.T) {
	p := IntPtr(42)

	if p == nil {
		t.Fatal("expected non-nil pointer, got nil")
	}
	if *p != 42 {
		t.Errorf("expected 42, got %d", *p)
	}
}

func TestIntPtr_ZeroValue(t *testing.T) {
	p := IntPtr(0)

	if p == nil {
		t.Fatal("expected non-nil pointer for zero, got nil")
	}
	if *p != 0 {
		t.Errorf("expected 0, got %d", *p)
	}
}

func TestIntPtr_NegativeValue(t *testing.T) {
	p := IntPtr(-100)

	if *p != -100 {
		t.Errorf("expected -100, got %d", *p)
	}
}

func TestIntPtr_DoesNotShareMemory(t *testing.T) {
	p1 := IntPtr(10)
	p2 := IntPtr(10)

	*p1 = 99

	if *p2 == 99 {
		t.Error("expected independent pointers; modifying p1 affected p2")
	}
}

func TestIntPtr_MaxInt(t *testing.T) {
	p := IntPtr(math.MaxInt)

	if *p != math.MaxInt {
		t.Errorf("expected %d, got %d", math.MaxInt, *p)
	}
}

func TestIntPtr_MinInt(t *testing.T) {
	p := IntPtr(math.MinInt)

	if *p != math.MinInt {
		t.Errorf("expected %d, got %d", math.MinInt, *p)
	}
}

// ---------------------------------------------------------------------------
// Int64Ptr
// ---------------------------------------------------------------------------

func TestInt64Ptr_ReturnsNonNullPointer(t *testing.T) {
	p := Int64Ptr(1234567890123)

	if p == nil {
		t.Fatal("expected non-nil pointer, got nil")
	}
	if *p != 1234567890123 {
		t.Errorf("expected 1234567890123, got %d", *p)
	}
}

func TestInt64Ptr_ZeroValue(t *testing.T) {
	p := Int64Ptr(0)

	if p == nil {
		t.Fatal("expected non-nil pointer for zero, got nil")
	}
	if *p != 0 {
		t.Errorf("expected 0, got %d", *p)
	}
}

func TestInt64Ptr_NegativeValue(t *testing.T) {
	p := Int64Ptr(-9999999999)

	if *p != -9999999999 {
		t.Errorf("expected -9999999999, got %d", *p)
	}
}

func TestInt64Ptr_MaxInt64(t *testing.T) {
	p := Int64Ptr(math.MaxInt64)

	if *p != math.MaxInt64 {
		t.Errorf("expected %d, got %d", math.MaxInt64, *p)
	}
}

func TestInt64Ptr_MinInt64(t *testing.T) {
	p := Int64Ptr(math.MinInt64)

	if *p != math.MinInt64 {
		t.Errorf("expected %d, got %d", math.MinInt64, *p)
	}
}

func TestInt64Ptr_DoesNotShareMemory(t *testing.T) {
	p1 := Int64Ptr(50)
	p2 := Int64Ptr(50)

	*p1 = -1

	if *p2 == -1 {
		t.Error("expected independent pointers; modifying p1 affected p2")
	}
}

// ---------------------------------------------------------------------------
// BoolPtr
// ---------------------------------------------------------------------------

func TestBoolPtr_True(t *testing.T) {
	p := BoolPtr(true)

	if p == nil {
		t.Fatal("expected non-nil pointer, got nil")
	}
	if *p != true {
		t.Errorf("expected true, got %v", *p)
	}
}

func TestBoolPtr_False(t *testing.T) {
	p := BoolPtr(false)

	if p == nil {
		t.Fatal("expected non-nil pointer for false, got nil")
	}
	if *p != false {
		t.Errorf("expected false, got %v", *p)
	}
}

func TestBoolPtr_DoesNotShareMemory(t *testing.T) {
	p1 := BoolPtr(true)
	p2 := BoolPtr(true)

	*p1 = false

	if *p2 == false {
		t.Error("expected independent pointers; modifying p1 affected p2")
	}
}

// ---------------------------------------------------------------------------
// Float64Ptr
// ---------------------------------------------------------------------------

func TestFloat64Ptr_ReturnsNonNullPointer(t *testing.T) {
	p := Float64Ptr(3.14)

	if p == nil {
		t.Fatal("expected non-nil pointer, got nil")
	}
	if *p != 3.14 {
		t.Errorf("expected 3.14, got %f", *p)
	}
}

func TestFloat64Ptr_ZeroValue(t *testing.T) {
	p := Float64Ptr(0.0)

	if p == nil {
		t.Fatal("expected non-nil pointer for 0.0, got nil")
	}
	if *p != 0.0 {
		t.Errorf("expected 0.0, got %f", *p)
	}
}

func TestFloat64Ptr_NegativeValue(t *testing.T) {
	p := Float64Ptr(-273.15)

	if *p != -273.15 {
		t.Errorf("expected -273.15, got %f", *p)
	}
}

func TestFloat64Ptr_VerySmallValue(t *testing.T) {
	p := Float64Ptr(1e-308)

	if *p != 1e-308 {
		t.Errorf("expected %e, got %e", 1e-308, *p)
	}
}

func TestFloat64Ptr_PositiveInfinity(t *testing.T) {
	p := Float64Ptr(math.Inf(1))

	if !math.IsInf(*p, 1) {
		t.Error("expected positive infinity")
	}
}

func TestFloat64Ptr_NegativeInfinity(t *testing.T) {
	p := Float64Ptr(math.Inf(-1))

	if !math.IsInf(*p, -1) {
		t.Error("expected negative infinity")
	}
}

func TestFloat64Ptr_NaN(t *testing.T) {
	p := Float64Ptr(math.NaN())

	if !math.IsNaN(*p) {
		t.Error("expected NaN")
	}
}

func TestFloat64Ptr_DoesNotShareMemory(t *testing.T) {
	p1 := Float64Ptr(1.5)
	p2 := Float64Ptr(1.5)

	*p1 = 99.9

	if *p2 == 99.9 {
		t.Error("expected independent pointers; modifying p1 affected p2")
	}
}

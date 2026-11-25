package photos

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsInvalidMediaItemID(t *testing.T) {
	missing := &ErrorResponse{Code: 400, Status: "INVALID_ARGUMENT", Message: "Invalid media item ID."}
	if !IsInvalidMediaItemID(missing) {
		t.Error("the observed missing-anchor error should classify as invalid media item ID")
	}
	otherStatus := &ErrorResponse{Code: 403, Status: "PERMISSION_DENIED", Message: "Invalid media item ID."}
	if IsInvalidMediaItemID(otherStatus) {
		t.Error("a different status should not classify as invalid media item ID")
	}
	otherMessage := &ErrorResponse{Code: 400, Status: "INVALID_ARGUMENT", Message: "Something else went wrong."}
	if IsInvalidMediaItemID(otherMessage) {
		t.Error("a different message should not classify as invalid media item ID")
	}
	if IsInvalidMediaItemID(fmt.Errorf("network error")) {
		t.Error("a plain error should not classify as invalid media item ID")
	}
	if IsInvalidMediaItemID(nil) {
		t.Error("a nil error should not classify as invalid media item ID")
	}
	wrapped := fmt.Errorf("add failed: %w", missing)
	if !IsInvalidMediaItemID(wrapped) {
		t.Error("a wrapped missing-anchor error should still classify")
	}
	var apiErr *ErrorResponse
	if !errors.As(missing, &apiErr) {
		t.Error("ErrorResponse should implement error so callers can unwrap it")
	}
}

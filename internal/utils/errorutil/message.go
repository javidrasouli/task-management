package errorutil

import "net/http"

var (
	ErrInternal         = ErrorType{Code: http.StatusInternalServerError, Message: "خطای داخلی سیستم لطفا مجددا تلاش فرمایید"}
	ErrNotFound         = ErrorType{Code: http.StatusNotFound, Message: "مورد مورد نظر یافت نشد"}
	ErrConflict         = ErrorType{Code: http.StatusConflict, Message: "این مورد از قبل وجود دارد"}
	ErrValidation       = ErrorType{Code: http.StatusUnprocessableEntity, Message: "اطلاعات ورودی معتبر نیست"}
	ErrBadRequest       = ErrorType{Code: http.StatusBadRequest, Message: "درخواست نامعتبر است"}
	ErrInvalidTransition = ErrorType{Code: http.StatusUnprocessableEntity, Message: "invalid status transition"}
)

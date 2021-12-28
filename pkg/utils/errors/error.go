package models

import (
	harbor2 "github.com/hchenc/go-harbor"
	git "github.com/xanzy/go-gitlab"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewConflict(err error) *errors.StatusError {
	if err == nil {
		return nil
	}
	switch t := err.(type) {
	case *git.ErrorResponse:
		return &errors.StatusError{ErrStatus: metav1.Status{
			Status:  metav1.StatusFailure,
			Code:    int32(t.Response.StatusCode),
			Reason:  metav1.StatusReasonConflict,
			Message: t.Message,
		},
		}
	case harbor2.GenericHarborError:
		return &errors.StatusError{ErrStatus: metav1.Status{
			Status:  metav1.StatusFailure,
			Reason:  metav1.StatusReasonConflict,
			Code:    int32(t.StatusCode()),
			Message: string(t.Body()),
		},
		}
	default:
		return &errors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Reason: metav1.StatusReasonUnknown,
		},
		}
	}
}

func NewNotFound(err error) *errors.StatusError {
	if err == nil {
		return nil
	}
	switch t := err.(type) {
	case *git.ErrorResponse:
		return &errors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Code:   int32(t.Response.StatusCode),
			Reason: metav1.StatusReasonNotFound,
		},
		}
	case *harbor2.GenericHarborError:
		return &errors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Reason: metav1.StatusReasonNotFound,
			Code:   int32(t.StatusCode()),
		},
		}
	default:
		return &errors.StatusError{ErrStatus: metav1.Status{
			Status: metav1.StatusFailure,
			Reason: metav1.StatusReasonUnknown,
		},
		}
	}
}

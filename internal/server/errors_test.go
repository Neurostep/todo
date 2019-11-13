package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/require"
)

func TestRespondErrors(t *testing.T) {
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = &http.Request{
		URL:    nil,
		Method: "GET",
	}

	respondErrors(c, log.NewNopLogger(), http.StatusBadRequest,
		newError("label1", "message1"),
		newError("label2", "message2"))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var errs Errors
	err := json.Unmarshal(rec.Body.Bytes(), &errs)
	require.NoError(t, err)

	require.Len(t, errs.Errors, 2)
	var e1, e2 *Error
	if errs.Errors[0].Label == "label1" {
		e1 = errs.Errors[0]
		e2 = errs.Errors[1]
	} else {
		e1 = errs.Errors[1]
		e2 = errs.Errors[0]
	}

	require.Equal(t, e1.Label, "label1")
	require.Equal(t, e1.Message, "message1")
	require.Equal(t, e2.Label, "label2")
	require.Equal(t, e2.Message, "message2")
}

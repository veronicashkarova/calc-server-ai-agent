///////////////////////////////////////////////////////////////////////
//
//  Don,t run tests by RunTests. Run every test independantly
//
///////////////////////////////////////////////////////////////////////

package application

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"github.com/veronicashkarova/server-for-calc/pkg/contract"
)

// Run this test independantly 
func TestExpressionsHandler(t *testing.T) {
	contract.ExpressionMap = map[string]contract.ExpressionMapData{}
	jsonRequest := `{"expression": "3+(8*3)"}`
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(jsonRequest))
	w := httptest.NewRecorder()
	NewExpressionHandler(w, req)
	req = httptest.NewRequest(http.MethodGet, "/", nil)
	w = httptest.NewRecorder()
	ExpressionsHandler(w,req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	var task contract.ExpressionsData
	err = json.Unmarshal(data, &task)
	if err != nil {
		t.Errorf("error get expressions")
	}

	if res.StatusCode != http.StatusOK{
		t.Errorf("wrong status code")
	}
}

// Run this test independantly
func TestNewExpressionHandler (t *testing.T) {

	contract.ExpressionMap = map[string]contract.ExpressionMapData{}
	jsonRequest := `{"expression": "3+(8*3)"}`
	req := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(jsonRequest))
	w := httptest.NewRecorder()
	NewExpressionHandler(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("Error: %v", err)
	}

	var task contract.ExpressionData
	err = json.Unmarshal(data, &task)
	if err != nil {
		t.Errorf("error get task")
	}

	if res.StatusCode != http.StatusCreated{
		t.Errorf("wrong status code")
	}
}


func TestIsIdExpressionRequest(t *testing.T) {
	stringUrl := "http://localhost/internal/task/1"
	parsedURL, err := url.Parse(stringUrl)
	if err == nil {
		id, err := isIdExpressionRequest(parsedURL)
		if err != nil || id != "1"  {
			t.Errorf("error get ID from url")
		}
	} 
}



package main

import (
	"fmt"
	"log"
	"reflect"
	"strings"
	"testing"
)

func TestDecompose(t *testing.T) {
	t.Run("test psn", func(t *testing.T) {
		ps, worker := decompose([]string{"h23", "h45", "h2", "h10"}, 2, "")
		expectedPs := []string{"h23", "h45"}
		expectedWorker := []string{"h2", "h10"}
		if !reflect.DeepEqual(ps, expectedPs){
			t.Errorf("expected: %v, Real:%v", expectedPs, ps)
		}

		if !reflect.DeepEqual(worker, expectedWorker){
			t.Errorf("expected: %v, Real: %v", expectedWorker, worker)
		}
	})
}

func TestFormatArr(t *testing.T) {
	arrStr := formatArr([]string{"h10", "h11", "h12"})
	expected := `"h10:2222","h11:2222","h12:2222"`
	if arrStr != expected {
		t.Errorf("Expected: %s, Real: %s", expected, arrStr)
	}
}

func TestFormatTtConfig(t *testing.T) {
	tConfig := formatTtConfig(`"h10:2222"`, `"h20:2222","h30:2222"`, 0, "ps")
	expected := `
	            {
                  "cluster": {
                    "ps": "h10:2222",
                    "worker": "h20:2222","h30:2222"
                  },
                  "task": {
                    "index": 0,
                    "type": "ps"
                  }
                }
`
	if strings.TrimSpace(tConfig) != strings.TrimSpace(expected){
		t.Errorf("Expected: %s, Real: %s", expected, tConfig)
	}
}

func TestTransfer(t *testing.T) {
	result := transfer("test_tmpl.yaml", []string{"h10"}, []string{"h11", "h12"})
	fmt.Println(result)
}

func TestResultFileName(t *testing.T) {
	expected := "test_tfjob.yaml"
	real := resultFileName("test.yaml")
	if real != expected{
		log.Fatalf("Expected: %s, Real: %s\n", expected, real)
	}

	real = resultFileName("test")
	if real != expected{
		log.Fatalf("Expected: %s, Real: %s", expected, real)
	}
}

func TestFilterNodesByExec(t *testing.T) {
	filtered := filterNodesByExec([]string{"h100", "h91", "h88", "h89"}, "exc")
	expected := []string{"h88", "h89"}
	if !reflect.DeepEqual(expected, filtered){
		t.Errorf("Expected: %v, Real: %v", expected, filtered)
	}
}

func TestFilerNodes(t *testing.T) {
	filtered := filterNodes([]string{"h100", "h91", "h88", "h89", "h202"}, true, 2)
	expected := []string{"h88", "h89"}
	if !reflect.DeepEqual(expected, filtered){
		t.Errorf("Expected: %v, Real: %v", expected, filtered)
	}
}
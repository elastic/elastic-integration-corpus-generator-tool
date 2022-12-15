package genlib

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/config"
	"github.com/elastic/elastic-integration-corpus-generator-tool/pkg/genlib/fields"
)

func Test_Gen2Plain(t *testing.T) {
	ctx := context.Background()
	// flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "endpoint", "process", "8.2.0")
	flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "aws", "vpcflow", "1.28.0")

	if err != nil {
		t.Fatal(err)
	}

	dat, err := os.ReadFile("testdata/template.tpl")
	if err != nil {
		t.Fatal(err)
	}

	g, err := NewGen2(dat, Config{}, flds)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	state := NewGenState()

	err = g.Emit(state, &buf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buf.String())
}

func Test_Gen2PlainWithConfig(t *testing.T) {
	ctx := context.Background()
	// flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "endpoint", "process", "8.2.0")
	flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "aws", "vpcflow", "1.28.0")

	if err != nil {
		t.Fatal(err)
	}

	dat, err := os.ReadFile("testdata/template.tpl")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadConfig("testdata/aws-vpcflow.conf.yaml")
	if err != nil {
		t.Fatal(err)
	}

	g, err := NewGen2(dat, cfg, flds)
	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	state := NewGenState()

	err = g.Emit(state, &buf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buf.String())
}

func Test_Gen2WithObjects(t *testing.T) {
	ctx := context.Background()
	flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "gcp", "gke", "2.15.1")

	if err != nil {
		t.Fatal(err)
	}

	dat, err := os.ReadFile("testdata/gcp-2.15.1.tpl")
	if err != nil {
		t.Fatal(err)
	}

	g, err := NewGen2(dat, Config{}, flds)

	if err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer

	state := NewGenState()

	err = g.Emit(state, &buf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buf.String())
}

// go test -bench=. -v -benchmem -count=20 -run "Benchmark_Gen2"
func Benchmark_Gen2(b *testing.B) {
	ctx := context.Background()
	// flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "endpoint", "process", "8.2.0")
	flds, err := fields.LoadFields(ctx, fields.ProductionBaseURL, "aws", "vpcflow", "1.28.0")

	if err != nil {
		b.Fatal(err)
	}

	dat, err := os.ReadFile("testdata/template.tpl")
	if err != nil {
		b.Fatal(err)
	}

	g, err := NewGen2(dat, Config{}, flds)

	if err != nil {
		b.Fatal(err)
	}

	var buf bytes.Buffer

	state := NewGenState()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := g.Emit(state, &buf)
		if err != nil {
			b.Fatal(err)
		}
		buf.Reset()
	}
}

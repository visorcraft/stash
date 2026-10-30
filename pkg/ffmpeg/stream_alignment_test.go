package ffmpeg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidHLSAlignment(t *testing.T) {
	tests := []struct {
		name      string
		stream    *StreamType
		alignment string
		want      bool
	}{
		{name: "empty HLS", stream: StreamTypeHLS, alignment: "", want: true},
		{name: "aligned HLS", stream: StreamTypeHLS, alignment: StreamAlignment16x8, want: true},
		{name: "invalid HLS", stream: StreamTypeHLS, alignment: "2x2", want: false},
		{name: "alignment on DASH", stream: StreamTypeDASHVideo, alignment: StreamAlignment16x8, want: false},
		{name: "empty DASH", stream: StreamTypeDASHVideo, alignment: "", want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidHLSAlignment(tt.stream, tt.alignment); got != tt.want {
				t.Fatalf("ValidHLSAlignment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStreamDirWithAlignment(t *testing.T) {
	const base = "scene_hls_720"
	if got := streamDirWithAlignment(base, ""); got != base {
		t.Fatalf("ordinary cache dir = %q", got)
	}
	if got := streamDirWithAlignment(base, StreamAlignment16x8); got != base+"_align16x8" {
		t.Fatalf("aligned cache dir = %q", got)
	}
}

func TestVideoFilterWithAlignment(t *testing.T) {
	ordinary := VideoFilter("scale=-2:720")
	if got := videoFilterWithAlignment(ordinary, ""); got != ordinary {
		t.Fatalf("ordinary filter = %q", got)
	}
	want := "scale=-2:720,pad=ceil(iw/16)*16:ceil(ih/8)*8:x=0:y=0:color=black"
	if got := videoFilterWithAlignment(ordinary, StreamAlignment16x8); string(got) != want {
		t.Fatalf("aligned filter = %q, want %q", got, want)
	}
}

func TestAlignedHLSForcesCPUAndDisablesFullHardware(t *testing.T) {
	codec := hlsCodecWithAlignment(nil, "hls", StreamAlignment16x8)
	if codec.CodeName != VideoCodecLibX264.CodeName {
		t.Fatalf("codec = %q, want libx264", codec.CodeName)
	}
	if fullHardwareWithAlignment(true, StreamAlignment16x8) {
		t.Fatal("aligned HLS must not use full hardware filters")
	}
	if !fullHardwareWithAlignment(true, "") {
		t.Fatal("ordinary HLS must preserve the hardware setting")
	}
}

func TestHLSSegmentQuery(t *testing.T) {
	req := httptest.NewRequest(
		http.MethodGet,
		"http://stash.example/scene/7/stream.m3u8?resolution=STANDARD&align=16x8&apikey=test-only",
		nil,
	)
	got := hlsSegmentQuery(req.URL, "STANDARD")
	want := "?align=16x8&apikey=test-only&resolution=STANDARD"
	if got != want {
		t.Fatalf("segment query = %q, want %q", got, want)
	}
}

func TestHLSSegmentQueryWithoutAlignment(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://stash.example/scene/7/stream.m3u8", nil)
	if got := hlsSegmentQuery(req.URL, ""); got != "" {
		t.Fatalf("ordinary segment query = %q, want empty", got)
	}
}

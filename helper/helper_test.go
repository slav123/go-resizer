package helper

import "testing"

var prepare_tests = []struct {
	url      string
	expected ImageOp
}{
	{"/core/cs3/w_280,h_190,c_fill/IMG_2934.JPG", ImageOp{280, 190, true, "IMG_2934.JPG", "44504e1b80bfcdf4b3d3a473568015a3", 0, 0, 0, 0, 0, 0, 85}},
	{"/core/cs3/h_180,w_190/IMG_2934.JPG", ImageOp{190, 180, false, "IMG_2934.JPG", "cb9f55b8a8669de3b740426aadf22895", 0, 0, 0, 0, 0, 0,85}},
	{"/core/cs3/h_180,w_190,t_jpg/IMG_2934.JPG", ImageOp{190, 180, false, "IMG_2934.JPG", "80a2f1110ce318fdd9be6046519cf6de", 1, 0, 0, 0, 0, 0,85}},
	{"/yll/pm/w_90,h_60,c_fill/Zachary%27s+Christening.jpg", ImageOp{90, 60, true, "Zachary's Christening.jpg", "aa44fc7e5d7810b68e9d4397ef2fecbf", 0, 0, 0, 0, 0, 0, 85}},
	{"/demo/cs3/w_200,h_201,t_100,l_101,x_40,y_41/image.jpg", ImageOp{200, 201, false, "image.jpg", "b58dde62efe1ba9a5ac68c70cbe22d3f", 0, 100, 101, 40, 41, 0, 85}},
	{"/demo/cs3/w_200,h_201,a_90/image.jpg", ImageOp{200, 201, false, "image.jpg", "94556faf4212fde526a9e15d223d72f7", 0, 0, 0, 0, 0, 90, 85}},
}

var test_ext = []struct {
	ext, expected string
}{
	{"png", "image/png"},
	{"jpg", "image/jpeg"},
	{"jpeg", "image/jpeg"},
	{"webp", "image/webp"},
	{"tiff", "image/tiff"},
	{"zzz", ""},
}

func TestPrepare(t *testing.T) {
	for _, mt := range prepare_tests {
		if v := Prepare(mt.url); v != mt.expected {
			t.Errorf("Prepare(%s) returned %v, expected %v", mt.url, v, mt.expected)
		}
	}
}

func TestImageType(t *testing.T) {
	for _, mt := range test_ext {
		if v := GetImageType(mt.ext); v != mt.expected {
			t.Errorf("GetImageType(%s) returned %v, expected %v", mt.ext, v, mt.expected)
		}
	}
}

func BenchmarkPrepare(b *testing.B) {

	//	var fn ImageOp

	for i := 0; i < b.N; i++ {
		_ = Prepare("/demo/s3/w_280,h_190,c_fill/IMG_2934.JPG")
	}
}

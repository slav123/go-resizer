package rwrap

import "testing"

func TestSetH(t *testing.T) {
	Server = "localhost"
	Port = "6379"
	Start()
	settting := SetH("test:Test", "Hash", "Value")
	if settting == false {
		t.Errorf("SetH failed")
	}
}

func TestHINCRBY(t *testing.T) {
	if HINCRBY("test:Test-inc", "incr", 2) == false {
		t.Errorf("HINCRBY")
	}
}

func TestIsSetH(t *testing.T) {
	if HExists("test:Test", "Hash") == false {
		t.Errorf("IsSetH")
	}
}

func TestGetH(t *testing.T) {
	val := GetH("test:Test", "Hash")
	if val != "Value" {
		t.Errorf("GetH failed")
	}
}

func TestMGetH(t *testing.T) {
	val := MGetH("test:Test")
	if val["Hash"] != "Value" {
		t.Errorf("GetH failed")
	}
}

func TestSetS(t *testing.T) {
	if !SetS("test:Single", "Value") {
		t.Errorf("Failed to SetS")
	}
}

func TestIsSet(t *testing.T) {
	if Exists("test:Single") == false {
		t.Errorf("Failed to IsSet")
	}
}

func TestGetS(t *testing.T) {
	if GetS("test:Single") != "Value" {
		t.Errorf("Failed to GetS")
	}
}

func TestMGet(t *testing.T) {

	result := MGet("test:Single", "Single2")

	if result[0] != "Value" {
		t.Errorf("Failed to MGet %v", result)
	}
}

func TestDel(t *testing.T) {

	result := Del("test:Single2")

	if !result {
		t.Errorf("Failed to Del")
	}
}

func TestHMSet(t *testing.T) {
	result := HMSet("test:HMSet", map[string]string{"a": "b", "c": "d"})
	if !result {
		t.Errorf("Failed to Del")
	}
}

func BenchmarkSetH(b *testing.B) {

	//	var fn ImageOp

	for i := 0; i < b.N; i++ {
		_ = GetH("Test", "Hash")
	}
}

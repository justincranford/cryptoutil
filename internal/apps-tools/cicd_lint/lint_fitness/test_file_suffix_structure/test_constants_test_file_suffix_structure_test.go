package test_file_suffix_structure

const (
	testBadRegex                                           = "[bad-regex"
	testFooBenchTestGo                                     = "foo_bench_test.go"
	testFooFuzzTestGo                                      = "foo_fuzz_test.go"
	testFooIntegrationTestGo                               = "foo_integration_test.go"
	testFooPropertyTestGo                                  = "foo_property_test.go"
	testFooTestGo                                          = "foo_test.go"
	testGoModNotFound                                      = "go.mod not found"
	testPackageMypkgImportTestingFuncBenchmarkFooBTestingB = "package mypkg\nimport \"testing\"\nfunc BenchmarkFoo(b *testing.B) {}\n"
	testPackageMypkgImportTestingFuncFuzzFooFTestingF      = "package mypkg\nimport \"testing\"\nfunc FuzzFoo(f *testing.F) {}\n"
	testPackageMypkgImportTestingFuncTestFooTTestingT      = "package mypkg\nimport \"testing\"\nfunc TestFoo(t *testing.T) {}\n"
	testViolationS                                         = "violation(s)"
)

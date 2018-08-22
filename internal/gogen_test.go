package internal

// func TestList(t *testing.T) {
// 	os.Setenv("GOGEN_FULLCONFIG", "")
// 	l := List()
// 	validateList(t, l)
// }

// func TestSearch(t *testing.T) {
// 	l := Search("weblog")
// 	validateList(t, l)
// }

// func TestGet(t *testing.T) {
// 	g, _ := Get("coccyx/weblog")
// 	assert.Equal(t, "coccyx/weblog", g.Gogen)
// 	assert.Equal(t, "weblog", g.Name)
// 	assert.Equal(t, "coccyx", g.Owner)
// }

// func TestUpsert(t *testing.T) {
// 	g, _ := Get("coccyx/weblog")
// 	Upsert(g)
// }

// func validateList(t *testing.T, l []GogenList) {
// 	if len(l) == 0 {
// 		t.Fatalf("Length of List() is 0")
// 	}
// 	if len(l[0].Gogen) == 0 {
// 		t.Fatalf("Gogen field of item 0 in List() is blank")
// 	}
// 	if len(l[0].Description) == 0 {
// 		t.Fatalf("Gogen description of item 0 in List() is blank")
// 	}
// }

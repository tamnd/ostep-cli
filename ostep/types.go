package ostep

// Chapter is one chapter from the OSTEP textbook.
type Chapter struct {
	Rank    int    `json:"rank"    csv:"rank"    tsv:"rank"`
	Chapter int    `json:"chapter" csv:"chapter" tsv:"chapter"`
	Part    string `json:"part"    csv:"part"    tsv:"part"`
	Title   string `json:"title"   csv:"title"   tsv:"title"`
	PDF     string `json:"pdf"     csv:"pdf"     tsv:"pdf"`
}

// SearchResult is one hit from a chapter-title search.
type SearchResult struct {
	Rank    int    `json:"rank"    csv:"rank"    tsv:"rank"`
	Chapter int    `json:"chapter" csv:"chapter" tsv:"chapter"`
	Part    string `json:"part"    csv:"part"    tsv:"part"`
	Title   string `json:"title"   csv:"title"   tsv:"title"`
	PDF     string `json:"pdf"     csv:"pdf"     tsv:"pdf"`
}

// Info is site-level stats.
type Info struct {
	Site     string `json:"site"     csv:"site"     tsv:"site"`
	Chapters int    `json:"chapters" csv:"chapters" tsv:"chapters"`
	Parts    int    `json:"parts"    csv:"parts"    tsv:"parts"`
	Source   string `json:"source"   csv:"source"   tsv:"source"`
}

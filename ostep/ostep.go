// Package ostep is the library behind the ostep CLI.
package ostep

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const DefaultUserAgent = "ostep-cli/dev (+https://github.com/tamnd/ostep-cli)"

type Config struct {
	BaseURL   string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
	UserAgent string
}

func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://pages.cs.wisc.edu/~remzi/OSTEP",
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
		UserAgent: DefaultUserAgent,
	}
}

type Client struct {
	cfg  Config
	http *http.Client
	last time.Time
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

var (
	cellRe  = regexp.MustCompile(`(?s)<td bgcolor=([^>]+)>(.*?)</td>`)
	numRe   = regexp.MustCompile(`<small>(\d+)</small>`)
	hrefRe  = regexp.MustCompile(`href=([\w./:-]+\.pdf)`)
	linkRe  = regexp.MustCompile(`href=[^>]+>([^<]+)</a>`)
	tagRe   = regexp.MustCompile(`<[^>]+>`)
)

var bgPart = map[string]string{
	"#f88017": "Virtualization",
	"#00aacc": "Concurrency",
	"#4cc417": "Persistence",
	"#3ea99f": "Security",
	"yellow":  "Intro",
}

// Chapters fetches the OSTEP home page and returns all numbered chapters.
func (c *Client) Chapters(ctx context.Context) ([]*Chapter, error) {
	body, err := c.get(ctx, c.cfg.BaseURL+"/")
	if err != nil {
		return nil, err
	}
	html := string(body)

	var chapters []*Chapter
	for _, m := range cellRe.FindAllStringSubmatch(html, -1) {
		bg := strings.ToLower(strings.Trim(m[1], `"' `))
		part, ok := bgPart[bg]
		if !ok {
			continue
		}
		content := m[2]

		nm := numRe.FindStringSubmatch(content)
		if nm == nil {
			continue
		}
		num, _ := strconv.Atoi(nm[1])

		lm := linkRe.FindStringSubmatch(content)
		if lm == nil {
			continue
		}
		title := strings.TrimSpace(tagRe.ReplaceAllString(lm[1], ""))

		hm := hrefRe.FindStringSubmatch(content)
		pdfFile := ""
		if hm != nil {
			pdfFile = hm[1]
		}
		pdfURL := ""
		if pdfFile != "" {
			pdfURL = c.cfg.BaseURL + "/" + pdfFile
		}

		chapters = append(chapters, &Chapter{
			Chapter: num,
			Part:    part,
			Title:   title,
			PDF:     pdfURL,
		})
	}

	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Chapter < chapters[j].Chapter
	})

	// Add rank after sort
	for i, ch := range chapters {
		ch.Rank = i + 1
	}

	return chapters, nil
}

// ChapterByNum fetches the chapter with the given number.
func (c *Client) ChapterByNum(ctx context.Context, num int) (*Chapter, error) {
	chapters, err := c.Chapters(ctx)
	if err != nil {
		return nil, err
	}
	for _, ch := range chapters {
		if ch.Chapter == num {
			return ch, nil
		}
	}
	return nil, fmt.Errorf("chapter %d not found", num)
}

// Search searches chapter titles for query (case-insensitive).
func (c *Client) Search(ctx context.Context, query string) ([]*SearchResult, error) {
	chapters, err := c.Chapters(ctx)
	if err != nil {
		return nil, err
	}
	q := strings.ToLower(query)
	var results []*SearchResult
	rank := 1
	for _, ch := range chapters {
		if strings.Contains(strings.ToLower(ch.Title), q) ||
			strings.Contains(strings.ToLower(ch.Part), q) {
			results = append(results, &SearchResult{
				Rank:    rank,
				Chapter: ch.Chapter,
				Part:    ch.Part,
				Title:   ch.Title,
				PDF:     ch.PDF,
			})
			rank++
		}
	}
	return results, nil
}

// Info returns site-level stats.
func (c *Client) Info(ctx context.Context) (*Info, error) {
	chapters, err := c.Chapters(ctx)
	if err != nil {
		return nil, err
	}
	parts := map[string]bool{}
	for _, ch := range chapters {
		parts[ch.Part] = true
	}
	return &Info{
		Site:     "pages.cs.wisc.edu/~remzi/OSTEP",
		Chapters: len(chapters),
		Parts:    len(parts),
		Source:   c.cfg.BaseURL,
	}, nil
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}

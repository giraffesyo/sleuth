package cnn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCNN(t *testing.T) {
	// serve up testdata/cnn
	testServer := httptest.NewServer(http.FileServer(http.Dir("testdata/cnn")))
	defer testServer.Close()
	t.Log("Test server started on ", testServer.URL)
	// run forever

	t.Run("Search", func(t *testing.T) {
		ctx := t.Context()
		baseUrl := testServer.URL + "/search.html?q="
		cnn := NewCNNProvider(ctx, WithCustomSearchUrl(baseUrl), WithoutPagination())
		videos, err := cnn.Search("body found")
		require.NoError(t, err)
		require.Len(t, videos, 9)

		require.Equal(t, "Passengers say cabin crew put a dead body next to them on flight", videos[0].Title)
		require.Equal(t, "https://www.cnn.com/2025/02/26/world/video/body-on-plane-qatar-airways-digvid", videos[0].Url)
		require.Equal(t, "Feb 26, 2025", videos[0].Date)
		t.Log(videos)

	})

}

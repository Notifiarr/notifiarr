package apps

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/Notifiarr/notifiarr/pkg/mnd"
	"github.com/gorilla/mux"
	"golift.io/qbit"
	"golift.io/starr"
	"golift.io/starr/debuglog"
)

type QbitConfig struct {
	ExtraConfig
	*qbit.Config
	*qbit.Qbit `toml:"-" xml:"-" json:"-"`
}

// qbitHandlers is called once on startup to register the web API paths.
func (a *Apps) qbitHandlers() {
	a.HandleAPIpath(Qbit, "category/set/{category}/{hash}", qbitSetCategory, "GET")
	a.HandleAPIpath(Qbit, "category/get", qbitGetCategory, "GET")
}

func getQbit(r *http.Request) *QbitConfig {
	return r.Context().Value(Qbit).(*QbitConfig) //nolint:forcetypeassert
}

func (a *Apps) setupQbit() error {
	for idx, app := range a.Qbit {
		if app == nil || app.Config == nil || app.URL == "" {
			return fmt.Errorf("%w: missing url: Qbit config %d", ErrInvalidApp, idx+1)
		} else if !strings.HasPrefix(app.Config.URL, "http://") && !strings.HasPrefix(app.Config.URL, "https://") {
			return fmt.Errorf("%w: URL must begin with http:// or https://: Qbit config %d", ErrInvalidApp, idx+1)
		}

		// a.Qbit[i].Debugf = a.DebugLog.Printf
		if err := a.Qbit[idx].Setup(a.MaxBody, a.Logger); err != nil {
			return err
		}
	}

	return nil
}

func (c *QbitConfig) Setup(maxBody int, logger mnd.Logger) error {
	if logger != nil && logger.DebugEnabled() {
		c.Client = starr.ClientWithDebug(c.Timeout.Duration, c.ValidSSL, debuglog.Config{
			MaxBody: maxBody,
			Debugf:  logger.Debugf,
			Caller:  metricMakerCallback("qBittorrent"),
			Redact:  []string{c.Pass, c.HTTPPass},
		})
	} else {
		c.Config.Client = starr.Client(c.Timeout.Duration, c.ValidSSL)
		c.Config.Client.Transport = NewMetricsRoundTripper("qBittorrent", c.Config.Client.Transport)
	}

	var err error
	if c.Qbit, err = qbit.NewNoAuth(c.Config); err != nil {
		return fmt.Errorf("qbit setup failed: %w", err)
	}

	return nil
}

// Enabled returns true if the instance is enabled and usable.
func (c *QbitConfig) Enabled() bool {
	return c != nil && c.Config != nil && c.URL != "" && c.Timeout.Duration >= 0
}

// @Description  Update the category for a torrent.
// @Summary      Set torrent category
// @Tags         Qbit
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Param        category  path   string  true  "category to set"
// @Param        hash  path   string  true  "torrent hash"
// @Success      200  {object} apps.Respond.apiResponse{message=string} "ok"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/qbit/{instance}/category/set/{category}/{hash} [get]
// @Security     ApiKeyAuth
func qbitSetCategory(req *http.Request) (int, interface{}) {
	category := mux.Vars(req)["category"]
	hash := mux.Vars(req)["hash"]

	err := getQbit(req).SetTorrentCategoryContext(req.Context(), category, hash)
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("setting category: %w", err)
	}

	return http.StatusOK, mnd.Success
}

// @Description  Return all configured categories.
// @Summary      Get all categories
// @Tags         Qbit
// @Produce      json
// @Param        instance  path   int64  true  "instance ID"
// @Success      200  {object} apps.Respond.apiResponse{message=[]string} "categories"
// @Failure      503  {object} apps.Respond.apiResponse{message=string} "instance error"
// @Failure      404  {object} string "bad token or api key"
// @Router       /api/qbit/{instance}/category/get [get]
// @Security     ApiKeyAuth
func qbitGetCategory(req *http.Request) (int, interface{}) {
	categories, err := getQbit(req).GetCategoriesContext(req.Context())
	if err != nil {
		return http.StatusServiceUnavailable, fmt.Errorf("getting categories: %w", err)
	}

	cats := []string{}
	for cat := range categories {
		cats = append(cats, cat)
	}

	return http.StatusOK, cats
}

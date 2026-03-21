package v4

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes は v4 のルートを登録する
func RegisterRoutes(rg *gin.RouterGroup) {
	// Basic 認証ミドルウェアで保護されたグループ
	authorized := rg.Group("/", gin.BasicAuth(gin.Accounts{
		"admin": "secret",
		"user":  "password",
	}))
	authorized.GET("/profile", getProfile)
	authorized.GET("/secret", getSecret)

	// goroutine を使った非同期処理のサンプル
	rg.GET("/async", asyncTasks)
}

// getProfile は Basic 認証後にユーザー情報を返す
// GET /api/v4/profile  (Authorization: Basic <base64(user:pass)>)
func getProfile(c *gin.Context) {
	// gin.BasicAuth が gin.AuthUserKey にログイン名をセットする
	user := c.MustGet(gin.AuthUserKey).(string)
	c.JSON(http.StatusOK, gin.H{
		"user":    user,
		"message": "authenticated successfully",
	})
}

// getSecret は Basic 認証後にのみアクセスできる機密リソースを返す
// GET /api/v4/secret
func getSecret(c *gin.Context) {
	user := c.MustGet(gin.AuthUserKey).(string)
	c.JSON(http.StatusOK, gin.H{
		"user":   user,
		"secret": "this is confidential",
	})
}

// asyncTasks は複数の goroutine を並列実行し、全結果を集約して返す
// GET /api/v4/async
func asyncTasks(c *gin.Context) {
	// gin.Context をコピーして goroutine 内で安全に使う
	cCp := c.Copy()

	type result struct {
		name     string
		value    string
		duration time.Duration
	}

	tasks := []struct {
		name  string
		sleep time.Duration
	}{
		{"task-a", 50 * time.Millisecond},
		{"task-b", 80 * time.Millisecond},
		{"task-c", 30 * time.Millisecond},
	}

	results := make([]result, len(tasks))
	var wg sync.WaitGroup

	for i, t := range tasks {
		wg.Add(1)
		go func(idx int, name string, sleep time.Duration) {
			defer wg.Done()
			start := time.Now()
			time.Sleep(sleep) // 非同期処理のシミュレーション
			elapsed := time.Since(start)
			log.Printf("[async] %s done (path: %s)", name, cCp.Request.URL.Path)
			results[idx] = result{name: name, value: name + " completed", duration: elapsed}
		}(i, t.name, t.sleep)
	}

	wg.Wait()

	out := make([]gin.H, len(results))
	for i, r := range results {
		out[i] = gin.H{
			"task":     r.name,
			"result":   r.value,
			"duration": r.duration.String(),
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"tasks": out,
	})
}

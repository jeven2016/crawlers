package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// https://www.cnblogs.com/YaoDD/p/11391316.html
var MetricsRuningCatalogPageTasksGauge prometheus.Gauge
var MetricsRuningNovelTasksGauge prometheus.Gauge
var MetricsRuningChapterTasksGauge prometheus.Gauge

var GaugeRuningTasks prometheus.Gauge

var MetricsTotalCatalogPageTasks prometheus.Counter
var MetricsTotalNovelTasks prometheus.Counter
var MetricsTotalChapterTasks prometheus.Counter
var MetricsComicPicDownloaded prometheus.Counter

var MetricsFailedCatalogPageTasksGauge prometheus.Gauge
var MetricsFailedNovelTasksGauge prometheus.Gauge
var MetricsFailedChapterTasksGauge prometheus.Gauge
var MetricsFailedComicPicTaskGauge prometheus.Gauge

var MetricsSucceedCatalogPageTasksGauge prometheus.Gauge
var MetricsSucceedNovelTasksGauge prometheus.Gauge
var MetricsSucceedChapterTasksGauge prometheus.Gauge

func init() {
	GaugeRuningTasks = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "running_tasks_gauge",
		Help: "running tasks gauge metrics",
	})

	MetricsRuningCatalogPageTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_running_catalog_page_tasks_count",
		Help: "The total number of running catalog page tasks",
	})

	MetricsRuningNovelTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_runing_novel_tasks_count",
		Help: "The total number of running novel tasks",
	})

	MetricsRuningChapterTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_runing_chapter_tasks_count",
		Help: "The total number of running chapter tasks",
	})

	MetricsTotalCatalogPageTasks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_total_catalog_page_tasks",
		Help: "The total number of chapter tasks",
	})

	MetricsTotalNovelTasks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_total_novel_tasks",
		Help: "The total number of novel tasks",
	})

	MetricsTotalChapterTasks = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_total_chapter_tasks",
		Help: "The total number of chapter tasks",
	})

	MetricsComicPicDownloaded = promauto.NewCounter(prometheus.CounterOpts{
		Name: "crawler_total_comic_picture_downloaded_tasks",
		Help: "The total number of comic pictures downloaded",
	})

	MetricsFailedCatalogPageTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_failed_catalog_page_tasks_count",
		Help: "The total number of failed catalog page tasks",
	})

	MetricsFailedNovelTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_failed_novel_tasks_count",
		Help: "The total number of failed novel tasks",
	})

	MetricsFailedChapterTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_failed_chapter_tasks_count",
		Help: "The total number of failed chapter tasks",
	})

	MetricsFailedComicPicTaskGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_failed_comic_pic_tasks_count",
		Help: "The total number of failed comic pic tasks",
	})

	MetricsSucceedCatalogPageTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_succeed_catalog_page_tasks_count",
		Help: "The total number of succeed catalog page tasks",
	})

	MetricsSucceedNovelTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_succeed_novel_tasks_count",
		Help: "The total number of succeed novel tasks",
	})

	MetricsSucceedChapterTasksGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "crawler_succeed_chapter_tasks_count",
		Help: "The total number of succeed chapter tasks",
	})
}

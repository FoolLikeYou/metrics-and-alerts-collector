package handler

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/Yandex-Practicum/go-musthave-metrics-tpl/internal/repository"
)

const listHTML = `<!DOCTYPE html>
<html lang="ru">
<head>
	<meta charset="utf-8">
	<title>Метрики</title>
</head>
<body>
	<h1>Метрики</h1>
	<table border="1" cellpadding="4" cellspacing="0">
		<thead>
			<tr><th>Имя</th><th>Тип</th><th>Значение</th></tr>
		</thead>
		<tbody>
		{{range .}}
			<tr><td>{{.Name}}</td><td>{{.MType}}</td><td>{{.Value}}</td></tr>
		{{end}}
		</tbody>
	</table>
</body>
</html>
`

var listTmpl = template.Must(template.New("metrics").Parse(listHTML))

// ListHTML отдаёт GET / — HTML со списком всех известных метрик.
func ListHTML(store repository.MetricsRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		rows := store.ListMetrics()
		var buf bytes.Buffer
		if err := listTmpl.Execute(&buf, rows); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = buf.WriteTo(w)
	}
}

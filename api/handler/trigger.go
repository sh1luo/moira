package handler

import (
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/moira-alert/moira-alert/api/controller"
	"github.com/moira-alert/moira-alert/api/dto"
	"net/http"
)

func trigger(router chi.Router) {
	router.Use(triggerContext)
	router.Put("/", saveTrigger)
	router.Get("/", getTrigger)
	router.Get("/state", getTriggerState)
	router.Route("/throttling", func(router chi.Router) {
		router.Get("/", getTriggerThrottling)
		router.Delete("/", deleteThrottling)
	})
	router.Route("/metrics", func(router chi.Router) {
		router.Get("/", getTriggerMetrics)
		router.Delete("/", deleteTriggerMetric)
	})
	router.Put("/maintenance", setMetricsMaintenance)
}

func saveTrigger(writer http.ResponseWriter, request *http.Request) {

}

func getTrigger(writer http.ResponseWriter, request *http.Request) {
	triggerId := request.Context().Value("triggerId").(string)
	trigger, err := controller.GetTrigger(database, triggerId)
	if err != nil {
		render.Render(writer, request, err)
		return
	}
	if err := render.Render(writer, request, trigger); err != nil {
		render.Render(writer, request, dto.ErrorRender(err))
	}
}

func getTriggerState(writer http.ResponseWriter, request *http.Request) {
	triggerId := request.Context().Value("triggerId").(string)
	triggerState, err := controller.GetTriggerState(database, triggerId)
	if err != nil {
		render.Render(writer, request, err)
		return
	}
	if err := render.Render(writer, request, triggerState); err != nil {
		render.Render(writer, request, dto.ErrorRender(err))
	}
}

func getTriggerThrottling(writer http.ResponseWriter, request *http.Request) {
	//not found
}

func deleteThrottling(writer http.ResponseWriter, request *http.Request) {
	//удалить throttling
}

func getTriggerMetrics(writer http.ResponseWriter, request *http.Request) {
	//not found
}

func deleteTriggerMetric(writer http.ResponseWriter, request *http.Request) {
	triggerId := request.Context().Value("triggerId").(string)
	metricName := request.URL.Query().Get("name")
	if metricName == "" {
		render.Render(writer, request, dto.ErrorInvalidRequest(fmt.Errorf("Metric name can not be empty")))
		return
	}

	if err := controller.DeleteTriggerMetric(database, metricName, triggerId); err != nil {
		render.Render(writer, request, err)
	}
}

func setMetricsMaintenance(writer http.ResponseWriter, request *http.Request) {
	triggerId := request.Context().Value("triggerId").(string)
	metricsMaintenance := &dto.MetricsMaintenance{}
	if err := render.Bind(request, metricsMaintenance); err != nil {
		render.Render(writer, request, dto.ErrorInvalidRequest(err))
		return
	}

	err := controller.SetMetricsMaintenance(database, triggerId, metricsMaintenance)
	if err != nil {
		render.Render(writer, request, err)
	}

	//Установить maintenance
	//в body - время, до которого будет maintenance
	//Умеет в массив треггеров
}
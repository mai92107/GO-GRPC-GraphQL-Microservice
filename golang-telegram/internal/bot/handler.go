package bot

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang-springboot-monitor-bot/internal/applog"
	"golang-springboot-monitor-bot/internal/chart"
	"golang-springboot-monitor-bot/internal/config"
	"golang-springboot-monitor-bot/internal/metric"
	"golang-springboot-monitor-bot/internal/model"
	"golang-springboot-monitor-bot/internal/repository"
	"golang-springboot-monitor-bot/internal/trend"
)

type Handler struct {
	repo     *repository.MemoryRepository
	trends   *trend.Repository
	renderer *chart.Renderer
	config   *config.Manager
	checkNow func(context.Context, string) error
	checkAll func(context.Context)
	sessions *SessionStore
	logger   applog.ApplicationLogger
}

const sessionTTL = 10 * time.Minute

func NewHandler(repo *repository.MemoryRepository) Handler {
	return Handler{repo: repo, sessions: NewSessionStore()}
}

func NewManagementHandler(repo *repository.MemoryRepository, trends *trend.Repository, renderer *chart.Renderer, manager *config.Manager, checkNow func(context.Context, string) error, checkAll func(context.Context), logger applog.ApplicationLogger) Handler {
	return Handler{repo: repo, trends: trends, renderer: renderer, config: manager, checkNow: checkNow, checkAll: checkAll, sessions: NewSessionStore(), logger: logger}
}

type Button struct {
	Text string `json:"text"`
	Data string `json:"callback_data"`
}

type Reply struct {
	Text     string
	Keyboard [][]Button
	Photo    []byte
	Silent   bool
}

func (handler Handler) HandleCommand(text string) string {
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return "請輸入指令。"
	}

	switch fields[0] {
	case "/list_services":
		return handler.listServices()
	case "/status":
		if len(fields) == 1 {
			return handler.statusAll()
		}
		return handler.status(fields[1])
	case "/metric":
		if len(fields) != 2 {
			return "用法：/metric <服務名稱>"
		}
		return handler.metrics(fields[1])
	case "/alerts":
		return handler.alerts()
	case "/menu":
		return "管理選單"
	case "/check":
		return "立即檢查只能透過已授權 Telegram 或執行中的管理介面使用。"
	case "/trend":
		return "趨勢圖只能透過已授權 Telegram 管理介面取得。"
	default:
		return "無法識別此指令。"
	}
}

func (handler Handler) Handle(ctx context.Context, chatID, text, callback string) Reply {
	if callback != "" {
		return handler.handleCallback(ctx, chatID, callback)
	}
	if session, ok := handler.sessions.Get(chatID); ok {
		return handler.handleSession(chatID, text, session)
	}
	fields := strings.Fields(strings.TrimSpace(text))
	if len(fields) == 0 {
		return Reply{Text: "請輸入指令。"}
	}
	switch fields[0] {
	case "/menu":
		return handler.menu()
	case "/metric":
		if len(fields) == 1 {
			return handler.metricServiceMenu()
		}
		return Reply{Text: handler.metrics(fields[1]), Keyboard: backKeyboard()}
	case "/trend":
		if len(fields) == 1 {
			return handler.trendServiceMenu()
		}
		return handler.trend(fields)
	case "/check":
		if len(fields) == 1 {
			if handler.checkAll == nil {
				return Reply{Text: "立即檢查目前不可用。"}
			}
			go handler.checkAll(ctx)
			return Reply{Text: "已開始檢查全部服務。"}
		}
		if handler.checkNow == nil {
			return Reply{Text: "立即檢查目前不可用。"}
		}
		if err := handler.checkNow(ctx, fields[1]); err != nil {
			return Reply{Text: err.Error()}
		}
		return Reply{Text: "已完成檢查 " + fields[1]}
	default:
		return Reply{Text: handler.HandleCommand(text)}
	}
}

func (handler Handler) menu() Reply {
	return Reply{Text: "管理選單", Keyboard: [][]Button{
		{{Text: "服務狀態", Data: "status"}, {Text: "目前告警", Data: "alerts"}},
		{{Text: "監控指標", Data: "metric_services"}, {Text: "趨勢圖", Data: "trend_services"}},
		{{Text: "服務管理", Data: "services"}},
		{{Text: "告警規則管理", Data: "rules"}, {Text: "立即檢查", Data: "check_all"}},
	}}
}

func (handler Handler) handleCallback(ctx context.Context, chatID, data string) Reply {
	switch data {
	case "status":
		return Reply{Text: handler.statusAll(), Keyboard: backKeyboard()}
	case "alerts":
		return handler.alertsWithButtons()
	case "trend_services":
		return handler.trendServiceMenu()
	case "metric_services":
		return handler.metricServiceMenu()
	case "services":
		return handler.serviceMenu()
	case "rules":
		return handler.ruleMenu()
	case "check_all":
		if handler.checkAll == nil {
			return Reply{Text: "立即檢查目前不可用。", Keyboard: backKeyboard()}
		}
		go handler.checkAll(ctx)
		return Reply{Text: "已開始檢查全部服務。", Keyboard: backKeyboard()}
	case "menu":
		return handler.menu()
	case "cancel":
		handler.sessions.Delete(chatID)
		return handler.menu()
	case "noop":
		return Reply{Silent: true}
	case "service_add":
		handler.startSession(chatID, "service_add", "name", nil)
		return Reply{Text: "請輸入服務名稱，或按取消。", Keyboard: cancelKeyboard()}
	case "rule_add":
		handler.startSession(chatID, "rule_add", "key", nil)
		return Reply{Text: "請輸入規則識別碼（key）。", Keyboard: cancelKeyboard()}
	}
	if strings.HasPrefix(data, "ack:") {
		id, err := strconv.ParseInt(strings.TrimPrefix(data, "ack:"), 10, 64)
		if err != nil {
			return Reply{Text: "無效的告警識別碼。", Keyboard: backKeyboard()}
		}
		if _, ok := handler.repo.AcknowledgeAlert(id, chatID, time.Now()); !ok {
			return Reply{Text: "告警不存在或已恢復。", Keyboard: backKeyboard()}
		}
		return Reply{Text: "已確認告警。", Keyboard: backKeyboard()}
	}
	if strings.HasPrefix(data, "metric_service:") {
		return handler.metricGroupMenu(strings.TrimPrefix(data, "metric_service:"))
	}
	if strings.HasPrefix(data, "metric_group:") {
		parts := strings.SplitN(strings.TrimPrefix(data, "metric_group:"), ":", 2)
		if len(parts) != 2 {
			return Reply{Text: "無效的指標分類。", Keyboard: backKeyboard()}
		}
		return handler.metricSubgroupMenu(parts[0], parts[1])
	}
	if strings.HasPrefix(data, "metric_subgroup:") {
		parts := strings.SplitN(strings.TrimPrefix(data, "metric_subgroup:"), ":", 3)
		if len(parts) != 3 {
			return Reply{Text: "無效的指標分類。", Keyboard: backKeyboard()}
		}
		return handler.metricsByCategory(parts[0], parts[1], parts[2])
	}
	if strings.HasPrefix(data, "ts:") {
		return handler.trendGroupMenu(strings.TrimPrefix(data, "ts:"))
	}
	if strings.HasPrefix(data, "tg:") {
		parts := strings.SplitN(strings.TrimPrefix(data, "tg:"), ":", 2)
		if len(parts) != 2 {
			return Reply{Text: "無效的指標分類。", Keyboard: backKeyboard()}
		}
		return handler.trendSubgroupMenu(parts[0], parts[1])
	}
	if strings.HasPrefix(data, "tu:") {
		parts := strings.SplitN(strings.TrimPrefix(data, "tu:"), ":", 3)
		if len(parts) != 3 {
			return Reply{Text: "無效的指標分類。", Keyboard: backKeyboard()}
		}
		return handler.trendMetricMenu(parts[0], parts[1], parts[2])
	}
	if strings.HasPrefix(data, "tm:") {
		parts := strings.SplitN(strings.TrimPrefix(data, "tm:"), ":", 2)
		if len(parts) != 2 {
			return Reply{Text: "無效的指標選擇。", Keyboard: backKeyboard()}
		}
		return handler.trendRangeMenu(parts[0], parts[1])
	}
	if strings.HasPrefix(data, "tr:") {
		parts := strings.SplitN(strings.TrimPrefix(data, "tr:"), ":", 3)
		if len(parts) != 3 {
			return Reply{Text: "無效的趨勢圖時間範圍。", Keyboard: backKeyboard()}
		}
		return handler.trend([]string{"/trend", parts[0], parts[1], parts[2]})
	}
	if data == "confirm_service_add" {
		return handler.applyServiceAdd(chatID)
	}
	if data == "confirm_rule_add" {
		return handler.applyRuleAdd(chatID)
	}
	if strings.HasPrefix(data, "service_edit:") {
		name := strings.TrimPrefix(data, "service_edit:")
		handler.startSession(chatID, "service_edit", "base_url", map[string]string{"name": name})
		return Reply{Text: "請輸入新的服務網址（Base URL）。", Keyboard: cancelKeyboard()}
	}
	if strings.HasPrefix(data, "rule_edit:") {
		key := strings.TrimPrefix(data, "rule_edit:")
		handler.startSession(chatID, "rule_edit", "threshold", map[string]string{"key": key})
		return Reply{Text: "請輸入新的門檻值（threshold）。", Keyboard: cancelKeyboard()}
	}
	if strings.HasPrefix(data, "rule_enable:") {
		return handler.ruleEnableMenu(chatID, strings.TrimPrefix(data, "rule_enable:"))
	}
	if strings.HasPrefix(data, "rule_enable_json:") {
		return handler.enableRuleWithJSONThreshold(chatID, strings.TrimPrefix(data, "rule_enable_json:"))
	}
	if strings.HasPrefix(data, "rule_enable_custom:") {
		key := strings.TrimPrefix(data, "rule_enable_custom:")
		handler.startSession(chatID, "rule_enable_custom", "threshold", map[string]string{"key": key})
		return Reply{Text: "請輸入自訂門檻值。此值將回寫 JSON 設定。", Keyboard: cancelKeyboard()}
	}
	if strings.HasPrefix(data, "rule_disable:") {
		key := strings.TrimPrefix(data, "rule_disable:")
		handler.startSession(chatID, "rule_disable:"+key, "confirm", nil)
		return Reply{Text: fmt.Sprintf("確認停用「%s」？", handler.ruleTitleByKey(key)), Keyboard: [][]Button{{{Text: "確認停用", Data: "confirm_mutation"}}, {{Text: "取消", Data: "cancel"}}}}
	}
	if data == "confirm_service_edit" || data == "confirm_rule_edit" {
		return handler.applyEdit(chatID)
	}
	if strings.HasPrefix(data, "service_toggle:") || strings.HasPrefix(data, "service_delete:") || strings.HasPrefix(data, "rule_delete:") {
		handler.startSession(chatID, data, "confirm", nil)
		return Reply{Text: "確認執行 " + data + "？", Keyboard: [][]Button{{{Text: "確認", Data: "confirm_mutation"}, {Text: "取消", Data: "cancel"}}}}
	}
	if data == "confirm_mutation" {
		return handler.applyMutation(chatID)
	}
	if data == "confirm_rule_enable_custom" {
		return handler.enableRuleWithCustomThreshold(chatID)
	}
	return Reply{Text: "未知操作。", Keyboard: backKeyboard()}
}

func (handler Handler) startSession(chatID, operation, step string, values map[string]string) {
	if values == nil {
		values = make(map[string]string)
	}
	handler.sessions.Put(model.TelegramSession{
		ChatID:    chatID,
		Operation: operation,
		Step:      step,
		Values:    values,
		ExpiresAt: time.Now().Add(sessionTTL),
	})
}

func (handler Handler) handleSession(chatID, text string, session model.TelegramSession) Reply {
	text = strings.TrimSpace(text)
	if session.Operation == "rule_add" {
		return handler.handleRuleAddSession(chatID, text, session)
	}
	if session.Operation == "service_edit" {
		switch session.Step {
		case "base_url":
			session.Values["base_url"] = text
			session.Step = "interval"
			handler.sessions.Put(session)
			return Reply{Text: "請輸入新的檢查間隔秒數。", Keyboard: cancelKeyboard()}
		case "interval":
			if _, err := strconv.Atoi(text); err != nil {
				return Reply{Text: "檢查間隔必須是整數秒。", Keyboard: cancelKeyboard()}
			}
			session.Values["interval"] = text
			session.Step = "confirm"
			handler.sessions.Put(session)
			return Reply{Text: fmt.Sprintf("修改服務 %s\n服務網址=%s\n檢查間隔=%s 秒", session.Values["name"], session.Values["base_url"], text), Keyboard: [][]Button{{{Text: "確認", Data: "confirm_service_edit"}, {Text: "取消", Data: "cancel"}}}}
		}
	}
	if session.Operation == "rule_edit" {
		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return Reply{Text: "門檻值（threshold）必須是數字。", Keyboard: cancelKeyboard()}
		}
		session.Values["threshold"] = text
		session.Step = "confirm"
		handler.sessions.Put(session)
		return Reply{Text: fmt.Sprintf("修改規則 %s\n門檻值=%s", session.Values["key"], text), Keyboard: [][]Button{{{Text: "確認", Data: "confirm_rule_edit"}, {Text: "取消", Data: "cancel"}}}}
	}
	if session.Operation == "rule_enable_custom" {
		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return Reply{Text: "自訂門檻值必須是數字。", Keyboard: cancelKeyboard()}
		}
		session.Values["threshold"] = text
		session.Step = "confirm"
		handler.sessions.Put(session)
		return Reply{
			Text:     fmt.Sprintf("確認啟用規則 %s，並將自訂門檻值 %s 回寫 JSON？", session.Values["key"], text),
			Keyboard: [][]Button{{{Text: "確認啟用", Data: "confirm_rule_enable_custom"}, {Text: "取消", Data: "cancel"}}},
		}
	}
	if session.Operation != "service_add" {
		return Reply{Text: "請使用確認或取消按鈕。", Keyboard: cancelKeyboard()}
	}
	switch session.Step {
	case "name":
		session.Values["name"] = text
		session.Step = "base_url"
		handler.sessions.Put(session)
		return Reply{Text: "請輸入服務網址（Base URL）。", Keyboard: cancelKeyboard()}
	case "base_url":
		session.Values["base_url"] = text
		session.Step = "interval"
		handler.sessions.Put(session)
		return Reply{Text: "請輸入檢查間隔秒數。", Keyboard: cancelKeyboard()}
	case "interval":
		if _, err := strconv.Atoi(text); err != nil {
			return Reply{Text: "檢查間隔必須是整數秒。", Keyboard: cancelKeyboard()}
		}
		session.Values["interval"] = text
		session.Step = "confirm"
		handler.sessions.Put(session)
		return Reply{Text: fmt.Sprintf("新增服務\n名稱=%s\n服務網址=%s\n檢查間隔=%s 秒", session.Values["name"], session.Values["base_url"], text), Keyboard: [][]Button{{{Text: "確認", Data: "confirm_service_add"}, {Text: "取消", Data: "cancel"}}}}
	default:
		return Reply{Text: "請使用確認或取消按鈕。", Keyboard: cancelKeyboard()}
	}
}

func (handler Handler) handleRuleAddSession(chatID, text string, session model.TelegramSession) Reply {
	next := map[string]string{"key": "service", "service": "metric", "metric": "operator", "operator": "threshold", "threshold": "severity"}
	prompts := map[string]string{"service": "請輸入服務名稱，套用全部服務請輸入 *。", "metric": "請輸入指標名稱（metric name）。", "operator": "請輸入比較運算子：>、>=、<、<=、==、!=。", "threshold": "請輸入門檻值（threshold）。", "severity": "請輸入嚴重程度（severity）。"}
	switch session.Step {
	case "metric":
		if err := metric.Validate(metric.Name(text)); err != nil {
			return Reply{Text: err.Error(), Keyboard: cancelKeyboard()}
		}
	case "operator":
		valid := map[string]bool{">": true, ">=": true, "<": true, "<=": true, "==": true, "!=": true}
		if !valid[text] {
			return Reply{Text: "比較運算子無效。", Keyboard: cancelKeyboard()}
		}
	case "threshold":
		if _, err := strconv.ParseFloat(text, 64); err != nil {
			return Reply{Text: "門檻值（threshold）必須是數字。", Keyboard: cancelKeyboard()}
		}
	}
	session.Values[session.Step] = text
	if session.Step == "severity" {
		session.Step = "confirm"
		handler.sessions.Put(session)
		return Reply{Text: fmt.Sprintf("新增規則\n識別碼=%s\n服務=%s\n指標=%s\n比較運算子=%s\n門檻值=%s\n嚴重程度=%s", session.Values["key"], session.Values["service"], session.Values["metric"], session.Values["operator"], session.Values["threshold"], text), Keyboard: [][]Button{{{Text: "確認", Data: "confirm_rule_add"}, {Text: "取消", Data: "cancel"}}}}
	}
	session.Step = next[session.Step]
	handler.sessions.Put(session)
	return Reply{Text: prompts[session.Step], Keyboard: cancelKeyboard()}
}

func (handler Handler) applyServiceAdd(chatID string) Reply {
	session, ok := handler.sessions.Get(chatID)
	if !ok || handler.config == nil {
		return Reply{Text: "操作已過期。", Keyboard: backKeyboard()}
	}
	interval, _ := strconv.Atoi(session.Values["interval"])
	handler.audit(chatID, "service_add_attempt", session.Values["name"])
	err := handler.config.Update(func(cfg *config.Config) error {
		cfg.Services = append(cfg.Services, config.Service{Name: session.Values["name"], BaseURL: session.Values["base_url"], CheckIntervalSeconds: interval, Enabled: true})
		return nil
	})
	if err != nil {
		return Reply{Text: "設定更新失敗：" + err.Error(), Keyboard: cancelKeyboard()}
	}
	handler.sessions.Delete(chatID)
	handler.audit(chatID, "service_add_completed", session.Values["name"])
	return Reply{Text: "服務已新增並立即生效。", Keyboard: backKeyboard()}
}

func (handler Handler) applyRuleAdd(chatID string) Reply {
	session, ok := handler.sessions.Get(chatID)
	if !ok || handler.config == nil {
		return Reply{Text: "操作已過期。", Keyboard: backKeyboard()}
	}
	threshold, _ := strconv.ParseFloat(session.Values["threshold"], 64)
	handler.audit(chatID, "rule_add_attempt", session.Values["key"])
	err := handler.config.Update(func(cfg *config.Config) error {
		cfg.AlertRules = append(cfg.AlertRules, config.AlertRule{Key: session.Values["key"], ServiceName: session.Values["service"], MetricName: metric.Name(session.Values["metric"]), Operator: session.Values["operator"], Threshold: threshold, Severity: session.Values["severity"], Enabled: true})
		return nil
	})
	if err != nil {
		return Reply{Text: "設定更新失敗：" + err.Error(), Keyboard: cancelKeyboard()}
	}
	handler.sessions.Delete(chatID)
	handler.audit(chatID, "rule_add_completed", session.Values["key"])
	return Reply{Text: "告警規則已新增並立即生效。", Keyboard: backKeyboard()}
}

func (handler Handler) applyEdit(chatID string) Reply {
	session, ok := handler.sessions.Get(chatID)
	if !ok || handler.config == nil {
		return Reply{Text: "操作已過期。", Keyboard: backKeyboard()}
	}
	handler.audit(chatID, session.Operation+"_attempt", session.Values["name"]+session.Values["key"])
	err := handler.config.Update(func(cfg *config.Config) error {
		switch session.Operation {
		case "service_edit":
			i, ok := config.FindService(*cfg, session.Values["name"])
			if !ok {
				return fmt.Errorf("service not found")
			}
			interval, _ := strconv.Atoi(session.Values["interval"])
			cfg.Services[i].BaseURL, cfg.Services[i].CheckIntervalSeconds = session.Values["base_url"], interval
		case "rule_edit":
			i, ok := config.FindAlertRule(*cfg, session.Values["key"])
			if !ok {
				return fmt.Errorf("rule not found")
			}
			threshold, _ := strconv.ParseFloat(session.Values["threshold"], 64)
			cfg.AlertRules[i].Threshold = threshold
		}
		return nil
	})
	if err != nil {
		return Reply{Text: "設定更新失敗：" + err.Error(), Keyboard: cancelKeyboard()}
	}
	handler.sessions.Delete(chatID)
	handler.audit(chatID, session.Operation+"_completed", session.Values["name"]+session.Values["key"])
	return Reply{Text: "設定已修改並立即生效。", Keyboard: backKeyboard()}
}

func (handler Handler) applyMutation(chatID string) Reply {
	session, ok := handler.sessions.Get(chatID)
	if !ok || handler.config == nil {
		return Reply{Text: "操作已過期。", Keyboard: backKeyboard()}
	}
	parts := strings.SplitN(session.Operation, ":", 2)
	handler.audit(chatID, parts[0]+"_attempt", parts[1])
	err := handler.config.Update(func(cfg *config.Config) error {
		switch parts[0] {
		case "service_toggle":
			i, ok := config.FindService(*cfg, parts[1])
			if !ok {
				return fmt.Errorf("service not found")
			}
			cfg.Services[i].Enabled = !cfg.Services[i].Enabled
		case "service_delete":
			i, ok := config.FindService(*cfg, parts[1])
			if !ok {
				return fmt.Errorf("service not found")
			}
			cfg.Services = append(cfg.Services[:i], cfg.Services[i+1:]...)
		case "rule_delete":
			i, ok := config.FindAlertRule(*cfg, parts[1])
			if !ok {
				return fmt.Errorf("rule not found")
			}
			cfg.AlertRules = append(cfg.AlertRules[:i], cfg.AlertRules[i+1:]...)
		case "rule_disable":
			i, ok := config.FindAlertRule(*cfg, parts[1])
			if !ok {
				return fmt.Errorf("找不到指定的告警規則")
			}
			cfg.AlertRules[i].Enabled = false
		}
		return nil
	})
	if err != nil {
		return Reply{Text: "設定更新失敗：" + err.Error(), Keyboard: cancelKeyboard()}
	}
	handler.sessions.Delete(chatID)
	handler.audit(chatID, parts[0]+"_completed", parts[1])
	return Reply{Text: "設定已更新並立即生效。", Keyboard: backKeyboard()}
}

func (handler Handler) ruleEnableMenu(chatID, key string) Reply {
	if handler.config == nil {
		return Reply{Text: "設定管理目前不可用。", Keyboard: backKeyboard()}
	}
	cfg := handler.config.Current()
	index, ok := config.FindAlertRule(cfg, key)
	if !ok {
		return Reply{Text: "找不到指定的告警規則。", Keyboard: backKeyboard()}
	}
	rule := cfg.AlertRules[index]
	return Reply{
		Text: fmt.Sprintf("啟用「%s」\nJSON 目前門檻值：%g\n請選擇門檻值來源：", ruleButtonTitle(rule), rule.Threshold),
		Keyboard: [][]Button{
			{{Text: fmt.Sprintf("使用 JSON 門檻值（%g）", rule.Threshold), Data: "rule_enable_json:" + key}},
			{{Text: "使用自訂門檻值", Data: "rule_enable_custom:" + key}},
			{{Text: "取消", Data: "cancel"}},
		},
	}
}

func (handler Handler) enableRuleWithJSONThreshold(chatID, key string) Reply {
	handler.audit(chatID, "rule_enable_json_attempt", key)
	err := handler.config.Update(func(cfg *config.Config) error {
		index, ok := config.FindAlertRule(*cfg, key)
		if !ok {
			return fmt.Errorf("找不到指定的告警規則")
		}
		cfg.AlertRules[index].Enabled = true
		return nil
	})
	if err != nil {
		return Reply{Text: "啟用規則失敗：" + err.Error(), Keyboard: backKeyboard()}
	}
	handler.audit(chatID, "rule_enable_json_completed", key)
	return Reply{Text: "規則已使用 JSON 目前門檻值啟用。", Keyboard: backKeyboard()}
}

func (handler Handler) enableRuleWithCustomThreshold(chatID string) Reply {
	session, ok := handler.sessions.Get(chatID)
	if !ok || handler.config == nil {
		return Reply{Text: "操作已過期。", Keyboard: backKeyboard()}
	}
	threshold, _ := strconv.ParseFloat(session.Values["threshold"], 64)
	key := session.Values["key"]
	handler.audit(chatID, "rule_enable_custom_attempt", key)
	err := handler.config.Update(func(cfg *config.Config) error {
		index, ok := config.FindAlertRule(*cfg, key)
		if !ok {
			return fmt.Errorf("找不到指定的告警規則")
		}
		cfg.AlertRules[index].Threshold = threshold
		cfg.AlertRules[index].Enabled = true
		return nil
	})
	if err != nil {
		return Reply{Text: "啟用規則失敗：" + err.Error(), Keyboard: cancelKeyboard()}
	}
	handler.sessions.Delete(chatID)
	handler.audit(chatID, "rule_enable_custom_completed", key)
	return Reply{Text: "自訂門檻值已回寫 JSON，規則已啟用。", Keyboard: backKeyboard()}
}

func (handler Handler) serviceMenu() Reply {
	reply := Reply{Text: "服務管理"}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "新增服務", Data: "service_add"}})
	for _, service := range handler.repo.Services() {
		toggleText := "啟用服務"
		if service.Enabled {
			toggleText = "停用服務"
		}
		reply.Keyboard = append(reply.Keyboard,
			[]Button{{Text: service.Name, Data: "noop"}},
			[]Button{
				{Text: toggleText, Data: "service_toggle:" + service.Name},
				{Text: "刪除", Data: "service_delete:" + service.Name},
			},
		)
	}
	reply.Keyboard = append(reply.Keyboard, backKeyboard()[0])
	return reply
}

func (handler Handler) ruleMenu() Reply {
	reply := Reply{Text: "告警規則管理"}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "新增規則", Data: "rule_add"}})
	if handler.config != nil {
		for _, rule := range handler.config.Current().AlertRules {
			toggle := Button{Text: "啟用規則", Data: "rule_enable:" + rule.Key}
			if rule.Enabled {
				toggle = Button{Text: "停用規則", Data: "rule_disable:" + rule.Key}
			}
			reply.Keyboard = append(reply.Keyboard,
				[]Button{{Text: ruleButtonTitle(rule), Data: "noop"}},
				[]Button{toggle, {Text: "刪除", Data: "rule_delete:" + rule.Key}},
			)
		}
	}
	reply.Keyboard = append(reply.Keyboard, backKeyboard()[0])
	return reply
}

func (handler Handler) ruleTitleByKey(key string) string {
	if handler.config == nil {
		return key
	}
	cfg := handler.config.Current()
	index, ok := config.FindAlertRule(cfg, key)
	if !ok {
		return key
	}
	return ruleButtonTitle(cfg.AlertRules[index])
}

func ruleButtonTitle(rule config.AlertRule) string {
	name := metricChineseNames[rule.MetricName.String()]
	if name == "" {
		name = rule.Key
	}
	return truncateButtonText(name, 24)
}

func truncateButtonText(value string, maxRunes int) string {
	runes := []rune(value)
	if len(runes) <= maxRunes {
		return value
	}
	if maxRunes <= 1 {
		return "…"
	}
	return string(runes[:maxRunes-1]) + "…"
}

func (handler Handler) alertsWithButtons() Reply {
	alerts := handler.repo.OpenAlerts()
	reply := Reply{Text: handler.alerts()}
	for _, alert := range alerts {
		if alert.AcknowledgedAt == nil {
			reply.Keyboard = append(reply.Keyboard, []Button{{Text: "確認 " + alert.ServiceName + "/" + alert.RuleKey, Data: fmt.Sprintf("ack:%d", alert.ID)}})
		}
	}
	reply.Keyboard = append(reply.Keyboard, backKeyboard()[0])
	return reply
}

func (handler Handler) metricServiceMenu() Reply {
	services := handler.repo.EnabledServices()
	if len(services) == 0 {
		return Reply{Text: "目前沒有啟用中的服務。", Keyboard: backKeyboard()}
	}
	reply := Reply{Text: "請選擇要查詢指標的服務："}
	for _, service := range services {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: service.Name, Data: "metric_service:" + service.Name}})
	}
	reply.Keyboard = append(reply.Keyboard, backKeyboard()[0])
	return reply
}

func (handler Handler) metricGroupMenu(serviceName string) Reply {
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	if len(metrics) == 0 {
		return Reply{Text: fmt.Sprintf("%s：目前尚無指標資料。", serviceName), Keyboard: backKeyboard()}
	}
	groups := metricCategories(metrics, 0, "")
	reply := Reply{Text: fmt.Sprintf("%s：請選擇指標大分類：", serviceName)}
	for _, group := range groups {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: metricGroupDisplayName(group), Data: "metric_group:" + serviceName + ":" + group}})
	}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "返回服務選擇", Data: "metric_services"}})
	return reply
}

func (handler Handler) metricSubgroupMenu(serviceName, group string) Reply {
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	subgroups := metricCategories(metrics, 1, group)
	if len(subgroups) == 0 {
		return Reply{Text: "此大分類目前沒有可用的細分類。", Keyboard: backKeyboard()}
	}
	reply := Reply{Text: fmt.Sprintf("%s / %s：請選擇指標細分類：", serviceName, metricGroupDisplayName(group))}
	for _, subgroup := range subgroups {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: metricSubgroupDisplayName(subgroup), Data: "metric_subgroup:" + serviceName + ":" + group + ":" + subgroup}})
	}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "返回大分類", Data: "metric_service:" + serviceName}})
	return reply
}

func (handler Handler) metricsByCategory(serviceName, group, subgroup string) Reply {
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	prefix := group + "_" + subgroup
	lines := []string{fmt.Sprintf("%s / %s / %s 指標資料：", serviceName, metricGroupDisplayName(group), metricSubgroupDisplayName(subgroup))}
	for _, snapshot := range metrics {
		if snapshot.Name == prefix || strings.HasPrefix(snapshot.Name, prefix+"_") {
			lines = append(lines, formatMetricSnapshot(snapshot))
		}
	}
	if len(lines) == 1 {
		return Reply{Text: "此分類目前沒有指標資料。", Keyboard: backKeyboard()}
	}
	return Reply{
		Text: strings.Join(lines, "\n\n"),
		Keyboard: [][]Button{
			{{Text: "返回細分類", Data: "metric_group:" + serviceName + ":" + group}},
			{{Text: "返回主選單", Data: "menu"}},
		},
	}
}

func (handler Handler) trendServiceMenu() Reply {
	services := handler.repo.EnabledServices()
	if len(services) == 0 {
		return Reply{Text: "目前沒有啟用中的服務。", Keyboard: backKeyboard()}
	}
	reply := Reply{Text: "請選擇要查詢趨勢圖的服務："}
	for _, service := range services {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: service.Name, Data: "ts:" + service.Name}})
	}
	reply.Keyboard = append(reply.Keyboard, backKeyboard()[0])
	return reply
}

func (handler Handler) trendGroupMenu(serviceName string) Reply {
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	if len(metrics) == 0 {
		return Reply{Text: fmt.Sprintf("%s：目前尚無指標資料。", serviceName), Keyboard: backKeyboard()}
	}
	reply := Reply{Text: fmt.Sprintf("%s：請選擇趨勢圖大分類：", serviceName)}
	for _, group := range metricCategories(metrics, 0, "") {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: metricGroupDisplayName(group), Data: "tg:" + serviceName + ":" + group}})
	}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "返回服務選擇", Data: "trend_services"}})
	return reply
}

func (handler Handler) trendSubgroupMenu(serviceName, group string) Reply {
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	reply := Reply{Text: fmt.Sprintf("%s / %s：請選擇趨勢圖細分類：", serviceName, metricGroupDisplayName(group))}
	for _, subgroup := range metricCategories(metrics, 1, group) {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: metricSubgroupDisplayName(subgroup), Data: "tu:" + serviceName + ":" + group + ":" + subgroup}})
	}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "返回大分類", Data: "ts:" + serviceName}})
	return reply
}

func (handler Handler) trendMetricMenu(serviceName, group, subgroup string) Reply {
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	names := metricNamesByCategory(metrics, group, subgroup)
	if len(names) == 0 {
		return Reply{Text: "此分類目前沒有可用的指標。", Keyboard: backKeyboard()}
	}
	reply := Reply{Text: fmt.Sprintf("%s / %s / %s：請選擇指標：", serviceName, metricGroupDisplayName(group), metricSubgroupDisplayName(subgroup))}
	for _, name := range names {
		reply.Keyboard = append(reply.Keyboard, []Button{{Text: metricDisplayName(name), Data: "tm:" + serviceName + ":" + name}})
	}
	reply.Keyboard = append(reply.Keyboard, []Button{{Text: "返回細分類", Data: "tg:" + serviceName + ":" + group}})
	return reply
}

func (handler Handler) trendRangeMenu(serviceName, metricName string) Reply {
	return Reply{
		Text: fmt.Sprintf("%s / %s：請選擇趨勢圖時間範圍：", serviceName, metricDisplayName(metricName)),
		Keyboard: [][]Button{
			{{Text: "最近 1 小時", Data: "tr:" + serviceName + ":" + metricName + ":1h"}},
			{{Text: "最近 4 小時", Data: "tr:" + serviceName + ":" + metricName + ":4h"}},
			{{Text: "最近 8 小時", Data: "tr:" + serviceName + ":" + metricName + ":8h"}},
			{{Text: "最近 16 小時", Data: "tr:" + serviceName + ":" + metricName + ":16h"}},
			{{Text: "最近 24 小時", Data: "tr:" + serviceName + ":" + metricName + ":24h"}},
			{{Text: "返回指標選擇", Data: "tu:" + serviceName + ":" + metricParts(metricName, 0) + ":" + metricParts(metricName, 1)}},
		},
	}
}

func metricNamesByCategory(metrics []model.MetricSnapshot, group, subgroup string) []string {
	values := make(map[string]struct{})
	for _, snapshot := range metrics {
		parts := strings.Split(snapshot.Name, "_")
		if len(parts) >= 2 && parts[0] == group && parts[1] == subgroup {
			values[snapshot.Name] = struct{}{}
		}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func metricDisplayName(name string) string {
	if chinese := metricChineseNames[name]; chinese != "" {
		return fmt.Sprintf("%s（%s）", chinese, name)
	}
	return name
}

func metricParts(name string, index int) string {
	parts := strings.Split(name, "_")
	if index < 0 || index >= len(parts) {
		return ""
	}
	return parts[index]
}

func metricCategories(metrics []model.MetricSnapshot, segment int, requiredGroup string) []string {
	values := make(map[string]struct{})
	for _, snapshot := range metrics {
		parts := strings.Split(snapshot.Name, "_")
		if len(parts) <= segment {
			continue
		}
		if requiredGroup != "" && parts[0] != requiredGroup {
			continue
		}
		values[parts[segment]] = struct{}{}
	}
	result := make([]string, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func metricGroupDisplayName(group string) string {
	names := map[string]string{
		"http":     "HTTP",
		"jvm":      "JVM",
		"hikaricp": "資料庫連線池",
		"process":  "程序",
		"system":   "系統",
	}
	if name := names[group]; name != "" {
		return fmt.Sprintf("%s（%s）", name, group)
	}
	return group
}

func metricSubgroupDisplayName(subgroup string) string {
	names := map[string]string{
		"server":      "伺服器",
		"memory":      "記憶體",
		"gc":          "垃圾回收",
		"threads":     "執行緒",
		"connections": "連線",
		"cpu":         "CPU",
		"uptime":      "運行時間",
	}
	if name := names[subgroup]; name != "" {
		return fmt.Sprintf("%s（%s）", name, subgroup)
	}
	return subgroup
}

func (handler Handler) trend(fields []string) Reply {
	if len(fields) != 4 || handler.trends == nil || handler.renderer == nil {
		return Reply{Text: "用法：/trend <service_name> <metric_name> <1h|4h|8h|16h|24h>"}
	}
	if err := metric.Validate(metric.Name(fields[2])); err != nil {
		return Reply{Text: err.Error()}
	}
	durations := map[string]time.Duration{"1h": time.Hour, "4h": 4 * time.Hour, "8h": 8 * time.Hour, "16h": 16 * time.Hour, "24h": 24 * time.Hour}
	duration, ok := durations[fields[3]]
	if !ok {
		return Reply{Text: "時間範圍必須是 1h、4h、8h、16h 或 24h。"}
	}
	samples, err := handler.trends.Query(fields[1], fields[2], duration, time.Now())
	if err != nil {
		return Reply{Text: err.Error()}
	}
	var thresholds []float64
	if handler.config != nil {
		for _, rule := range handler.config.Current().AlertRules {
			if rule.Enabled && (rule.ServiceName == "*" || rule.ServiceName == fields[1]) && rule.MetricName.String() == fields[2] {
				thresholds = append(thresholds, rule.Threshold)
			}
		}
	}
	var out bytes.Buffer
	if err := handler.renderer.Render(&out, fields[1], fields[2], fields[3], samples, thresholds); err != nil {
		return Reply{Text: err.Error()}
	}
	return Reply{Text: fmt.Sprintf("%s %s 最近 %s", fields[1], fields[2], fields[3]), Photo: out.Bytes()}
}

func backKeyboard() [][]Button   { return [][]Button{{{Text: "返回主選單", Data: "menu"}}} }
func cancelKeyboard() [][]Button { return [][]Button{{{Text: "取消", Data: "cancel"}}} }

func (handler Handler) audit(chatID, event, target string) {
	if handler.logger != nil {
		handler.logger.Info(context.Background(), event,
			applog.Field{Key: "component", Value: "telegram_management"},
			applog.Field{Key: "chat_id", Value: chatID},
			applog.Field{Key: "target", Value: target},
		)
	}
}

func (handler Handler) listServices() string {
	services := handler.repo.Services()
	if len(services) == 0 {
		return "目前沒有服務。"
	}

	lines := make([]string, 0, len(services))
	for _, service := range services {
		state := "停用"
		if service.Enabled {
			state = "啟用"
		}
		lines = append(lines, fmt.Sprintf(
			"名稱=%s \n環境=%s \n狀態=%s \nURL=%s \n檢查間隔=%d 秒\n",
			service.Name,
			service.Environment,
			state,
			service.BaseURL,
			service.CheckIntervalSeconds,
		))
	}
	return strings.Join(lines, "\n")
}

func (handler Handler) statusAll() string {
	services := handler.repo.EnabledServices()
	if len(services) == 0 {
		return "目前沒有啟用中的服務。"
	}

	lines := make([]string, 0, len(services))
	for _, service := range services {
		lines = append(lines, handler.statusLine(service.Name))
	}
	return strings.Join(lines, "\n\n")
}

func (handler Handler) status(serviceName string) string {
	if _, ok := handler.repo.Service(serviceName); !ok {
		return fmt.Sprintf("找不到服務 %q。", serviceName)
	}
	return handler.statusLine(serviceName)
}

func (handler Handler) statusLine(serviceName string) string {
	check, ok := handler.repo.LastHealthCheck(serviceName)
	if !ok {
		return fmt.Sprintf("%s：尚未檢查", serviceName)
	}
	return fmt.Sprintf("%s：%s", check.ServiceName, check.Status)
}

func (handler Handler) metrics(serviceName string) string {
	if _, ok := handler.repo.Service(serviceName); !ok {
		return fmt.Sprintf("找不到服務 %q。", serviceName)
	}
	metrics := handler.repo.LastMetricSnapshots(serviceName)
	if len(metrics) == 0 {
		return fmt.Sprintf("%s：目前尚無指標資料。", serviceName)
	}

	lines := []string{serviceName + " 指標資料："}
	for _, snapshot := range metrics {
		lines = append(lines, formatMetricSnapshot(snapshot))
	}
	return strings.Join(lines, "\n\n")
}

func formatLabels(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}

	names := make([]string, 0, len(labels))
	for name := range labels {
		names = append(names, name)
	}
	sort.Strings(names)

	pairs := make([]string, 0, len(names))
	for _, name := range names {
		pairs = append(pairs, fmt.Sprintf("%s=%q", name, labels[name]))
	}
	return "{" + strings.Join(pairs, ",") + "}"
}

func (handler Handler) alerts() string {
	alerts := handler.repo.OpenAlerts()
	if len(alerts) == 0 {
		return "目前沒有未恢復的告警。"
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].ServiceName < alerts[j].ServiceName
	})

	lines := make([]string, 0, len(alerts))
	for _, event := range alerts {
		lines = append(lines, formatAlert(event))
	}
	return strings.Join(lines, "\n")
}

func formatAlert(event model.AlertEvent) string {
	return fmt.Sprintf("%s %s\n", event.Severity, event.Message)
}

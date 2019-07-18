package chrome

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/ecwid/cdp"
)

const (
	// DriverMaxRetrySeconds максимальное время, в течение которого можно выполять ретрай
	DriverMaxRetrySeconds = 60
	// DriverRetryDelayMilliseconds пауза между попытками ретрая
	DriverRetryDelayMilliseconds = 500
)

// Тип картинок при снятии скрина
const (
	PNG  = "png"
	JPEG = "jpeg"
)

// Cookie ...
type Cookie = cdp.CookieParam

// Driver ...
type Driver struct {
	PanicInterceptor func(error)
	Logger           func(string, []interface{}, []interface{})
	session          *cdp.Session
}

// Session current session associated with Driver
func (e *Driver) Session() *cdp.Session {
	return e.session
}

// ErrDriver фатальная ошибка драйвера
type ErrDriver struct {
	SessionID string
	Message   string
}

func (e ErrDriver) Error() string {
	return fmt.Sprintf("sessionID: %s, error: %s", e.SessionID, e.Message)
}

// NewDriver ...
func NewDriver(client *cdp.Client) *Driver {
	return &Driver{session: client.NewSession(nil)}
}

func (e *Driver) log(params []interface{}, result []interface{}) {
	if e.Logger == nil {
		return
	}
	pc, _, _, _ := runtime.Caller(1)
	name := runtime.FuncForPC(pc).Name()
	name = name[strings.LastIndex(name, ".")+1 : len(name)]
	e.Logger(name, params, result)
}

// Panic ...
func (e *Driver) Panic(format string, v ...interface{}) {
	var err = &ErrDriver{
		SessionID: e.session.ID(),
		Message:   fmt.Sprintf(format, v...),
	}
	if e.PanicInterceptor != nil {
		e.PanicInterceptor(err)
	} else {
		panic(err)
	}
}

func retry(retry func() error) error {
	bk := &backoff{
		Max:   time.Second * DriverMaxRetrySeconds,
		Delay: time.Millisecond * DriverRetryDelayMilliseconds,
		Retry: retry,
	}
	return bk.Do()
}

// CurrentURL возвращает текущий URL вкладки
func (e *Driver) CurrentURL() string {
	var url string
	var err error
	_ = retry(func() error {
		url, err = e.session.URL()
		return err
	})
	e.log(nil, []interface{}{url})
	if err != nil {
		e.Panic(`Не могу получить URL текущей страницы: %s`, err)
	}
	return url
}

// Close закрывает текущую вкладку браузера (сессия тоже закрывается и ей больше нельзя пользоваться)
func (e *Driver) Close() {
	_, err := e.session.Close()
	e.log(nil, nil)
	if err != nil {
		e.Panic(`Не могу закрыть текущую вкладку браузера: %s`, err)
	}
}

// GetScreenshot делает скриин текущей страницы
func (e *Driver) GetScreenshot(format string, quality int8, full bool) []byte {
	img, _ := e.session.GetScreenshot(format, quality, full)
	return img
}

// MainFrame переключаемся на главный фрейм страницы (сама страница)
func (e *Driver) MainFrame() {
	e.session.MainFrame()
	e.log(nil, nil)
}

// Reload рефрешит текущую страницу
func (e *Driver) Reload() {
	err := retry(func() error {
		return e.session.Reload()
	})
	e.log(nil, nil)
	if err != nil {
		e.Panic("Reload страницы не сработал после всех попыток: %s", err)
	}
}

// SetHeaders ...
func (e *Driver) SetHeaders(headers map[string]string) {
	e.session.SetExtraHTTPHeaders(headers)
	e.log([]interface{}{headers}, nil)
}

// Script выполняет js код
func (e *Driver) Script(code string) (interface{}, error) {
	r, err := e.session.Script(code)
	e.log([]interface{}{code}, []interface{}{r, err})
	return r, err
}

// Navigate навигейтим текущую вкладку на адрес url
func (e *Driver) Navigate(url string) {
	err := retry(func() error {
		return e.session.Navigate(url)
	})
	e.log([]interface{}{url}, nil)
	if err != nil {
		e.Panic("Navigate на [%s] не сработал после всех попыток: %s", url, err)
	}
}

// SetCookies устанавливаем куки браузеру
func (e *Driver) SetCookies(cookies ...Cookie) {
	e.session.SetCookies(cookies...)
	e.log([]interface{}{cookies}, nil)
}

// SwitchToFrame переключается на фрейм по селектору элемента
func (e *Driver) SwitchToFrame(selector string) {
	err := retry(func() error {
		return e.session.SwitchToFrame(selector)
	})
	e.log([]interface{}{selector}, nil)
	if err != nil {
		e.Panic("Не могу переключиться на фрейм [%s]: %s", selector, err)
	}
}

// NewDriver инициирует новую сессию с указанным окном браузера
func (e *Driver) NewDriver(targetID string) *Driver {
	// Создаем новую сессию драйвера
	driver := &Driver{
		PanicInterceptor: e.PanicInterceptor,
		Logger:           e.Logger,
		session:          e.session.Client().NewSession(&targetID),
	}
	e.log([]interface{}{targetID}, nil)
	return driver
}

// GetComputedStyle ...
func (e *Driver) GetComputedStyle(selector string, style string) string {
	var comp = ""
	var err error
	_ = retry(func() error {
		comp, err = e.session.GetComputedStyle(selector, style)
		return err
	})
	e.log([]interface{}{selector, style}, []interface{}{comp})
	if err != nil {
		e.Panic("Не могу определить значение стиля [%v] для элемента %s: %s", style, selector, err)
	}
	return comp
}

// Select выбирает из selector опции args (поддерживает множественный выбор)
func (e *Driver) Select(selector string, args ...string) {
	err := retry(func() error {
		return e.session.Select(selector, args...)
	})
	e.log([]interface{}{selector, args}, nil)
	if err != nil {
		e.Panic("Не могу выбрать опции [%v] для select элемента %s: %s", args, selector, err)
	}
}

// Upload загружает файл в форму
func (e *Driver) Upload(selector string, file string) {
	err := retry(func() error {
		return e.session.Upload(selector, file)
	})
	e.log([]interface{}{selector, file}, nil)
	if err != nil {
		e.Panic("Не могу загрузить файл в элемент %s: %s", selector, err)
	}
}

func (e *Driver) getSelected(selector string, selectedText bool) []string {
	var text []string
	var err error
	_ = retry(func() error {
		text, err = e.session.GetSelected(selector, selectedText)
		return err
	})
	e.log([]interface{}{selector, selectedText}, []interface{}{text})
	if err != nil {
		e.Panic("Не могу получить выбранную опцию из select элемента %s: %s", selector, err)
	}
	return text
}

// GetSelected получает value выбранной опции из select
func (e *Driver) GetSelected(selector string) []string {
	return e.getSelected(selector, false)
}

// GetSelectedText получает текст выбранной опции из select
func (e *Driver) GetSelectedText(selector string) []string {
	return e.getSelected(selector, true)
}

// Click по элементу selector
func (e *Driver) Click(selector string) {
	err := retry(func() error {
		return e.session.Click(selector, nil)
	})
	e.log([]interface{}{selector}, nil)
	if err != nil {
		e.Panic("Не могу кликнуть по элементу %s: %s", selector, err)
	}
}

// Checkbox ...
func (e *Driver) Checkbox(selector string, check bool) {
	err := retry(func() error {
		return e.session.Checkbox(selector, check)
	})
	e.log([]interface{}{selector, check}, nil)
	if err != nil {
		e.Panic("Не могу установить checkbox в [%t] для элемента %s: %s", check, selector, err)
	}
}

// IsChecked ...
func (e *Driver) IsChecked(selector string) bool {
	var checked bool
	var err error
	_ = retry(func() error {
		checked, err = e.session.IsChecked(selector)
		return err
	})
	e.log([]interface{}{selector}, []interface{}{checked})
	if err != nil {
		e.Panic("Не могу определить, установлен ли checkbox для элемента %s: %s", selector, err)
	}
	return checked
}

// SetAttr ...
func (e *Driver) SetAttr(selector string, attr string, value string) {
	err := retry(func() error {
		return e.session.SetAttr(selector, attr, value)
	})
	e.log([]interface{}{selector, attr, value}, nil)
	if err != nil {
		e.Panic("Не могу установить значение аттрибута [%s = %s] для элемента `%s`: %s", attr, value, selector, err)
	}
}

// GetAttr ...
func (e *Driver) GetAttr(selector string, attr string) string {
	var value string
	var err error
	_ = retry(func() error {
		value, err = e.session.GetAttr(selector, attr)
		return err
	})
	e.log([]interface{}{selector, attr}, []interface{}{value})
	if err != nil {
		e.Panic("Не могу получить значение аттрибута [%s] для элемента `%s`: %s", attr, selector, err)
	}
	return value
}

// Type вставляет текст text в элемент selector и вводит клавиши keys
func (e *Driver) Type(selector string, text string, keys ...rune) {
	err := retry(func() error {
		err := e.session.Type(selector, text, keys...)
		if err != nil {
			return err
		}
		var gettext []string
		gettext, err = e.session.Text(selector)
		if text != gettext[0] {
			return fmt.Errorf("Введенный текст не совпадает с вводимым текстом exp: `%s`, was: `%s`", text, gettext[0])
		}
		return err
	})
	e.log([]interface{}{selector, text, keys}, nil)
	if err != nil {
		e.Panic("Не могу вписать текст в элемент %s: %#v", selector, err)
	}
}

// Predicate ожидает выполнение условия функции predicate
func (e *Driver) Predicate(text string, predicate func() bool) {
	err := retry(func() error {
		if predicate() {
			return nil
		}
		return errors.New("")
	})
	e.log([]interface{}{text}, nil)
	if err != nil {
		e.Panic("Ожидаемое событие [%s] не произошло", text)
	}
}

// Exist ожидает присутствие любого из переданных selector-ов в текущем Frame
// возвращает найденный селектор
func (e *Driver) Exist(selector ...string) string {
	m := fmt.Sprintf("[%s] должен появиться", strings.Join(selector, " или "))
	var exist = ""
	e.Predicate(m, func() bool {
		for _, s := range selector {
			if e.session.Count(s) > 0 {
				exist = s
				return true
			}
		}
		return false
	})
	e.log([]interface{}{selector}, []interface{}{exist})
	return exist
}

// NotExist ожидает отсутствия элемента selector в текущем Frame
func (e *Driver) NotExist(selector string) {
	e.Predicate(selector+" должен исчезнуть", func() bool {
		return e.session.Count(selector) == 0
	})
}

// Closed ждем, пока текущее окно браузера закроется (сессией после этого пользоваться нельзя)
func (e *Driver) Closed() {
	e.Predicate("окно должно закрыться", func() bool {
		return e.session.IsClosed()
	})
}

// Text получает текст элемента selector
func (e *Driver) Text(selector string) string {
	texts := e.TextAll(selector)
	if texts != nil && len(texts) > 0 {
		return texts[0]
	}
	return ""
}

// TextAll получает текст всех элементов, найденных по selector
func (e *Driver) TextAll(selector string) []string {
	var text []string
	var err error
	_ = retry(func() error {
		text, err = e.session.Text(selector)
		return err
	})
	e.log([]interface{}{selector}, []interface{}{text})
	if err != nil {
		e.Panic("Не могу получить текст из элементов %s: %s", selector, err)
	}
	return text
}

// NewTarget ожидает открытие новой вкладки
func (e *Driver) NewTarget(init func()) string {
	targetInfo := e.session.TargetCreated()
	e.log(nil, nil)
	init()
	select {
	case target := <-targetInfo:
		return target.TargetID
	case <-time.After(cdp.NavigationTimeout):
		panic("Новая вкладка браузера (попап окно) не открылась, как ожидали")
	}
}

// Count число элементов selector в текущем фрейме
func (e *Driver) Count(selector string) int {
	var count int
	_ = retry(func() error {
		count = e.session.Count(selector)
		return nil
	})
	e.log([]interface{}{selector}, []interface{}{count})
	return count
}


/*
 * generated by event_generator
 *
 * DO NOT EDIT
 */

package web

import "github.com/joernweissenborn/eventual2go"



type messageCompleter struct {
	*eventual2go.Completer
}

func NewmessageCompleter() *messageCompleter {
	return &messageCompleter{eventual2go.NewCompleter()}
}

func (c *messageCompleter) Complete(d message) {
	c.Completer.Complete(d)
}

func (c *messageCompleter) Future() *messageFuture {
	return &messageFuture{c.Completer.Future()}
}

type messageFuture struct {
	*eventual2go.Future
}

func (f *messageFuture) Result() message {
	return f.Future.Result().(message)
}

type messageCompletionHandler func(message) message

func (ch messageCompletionHandler) toCompletionHandler() eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		return ch(d.(message))
	}
}

func (f *messageFuture) Then(ch messageCompletionHandler) *messageFuture {
	return &messageFuture{f.Future.Then(ch.toCompletionHandler())}
}

func (f *messageFuture) AsChan() chan message {
	c := make(chan message, 1)
	cmpl := func(d chan message) messageCompletionHandler {
		return func(e message) message {
			d <- e
			close(d)
			return e
		}
	}
	ecmpl := func(d chan message) eventual2go.ErrorHandler {
		return func(error) (eventual2go.Data, error) {
			close(d)
			return nil, nil
		}
	}
	f.Then(cmpl(c))
	f.Err(ecmpl(c))
	return c
}

type messageStreamController struct {
	*eventual2go.StreamController
}

func NewmessageStreamController() *messageStreamController {
	return &messageStreamController{eventual2go.NewStreamController()}
}

func (sc *messageStreamController) Add(d message) {
	sc.StreamController.Add(d)
}

func (sc *messageStreamController) Join(s *messageStream) {
	sc.StreamController.Join(s.Stream)
}

func (sc *messageStreamController) JoinFuture(f *messageFuture) {
	sc.StreamController.JoinFuture(f.Future)
}

func (sc *messageStreamController) Stream() *messageStream {
	return &messageStream{sc.StreamController.Stream()}
}

type messageStream struct {
	*eventual2go.Stream
}

type messageSubscriber func(message)

func (l messageSubscriber) toSubscriber() eventual2go.Subscriber {
	return func(d eventual2go.Data) { l(d.(message)) }
}

func (s *messageStream) Listen(ss messageSubscriber) *eventual2go.Completer {
	return s.Stream.Listen(ss.toSubscriber())
}

func (s *messageStream) ListenNonBlocking(ss messageSubscriber) *eventual2go.Completer {
	return s.Stream.ListenNonBlocking(ss.toSubscriber())
}

type messageFilter func(message) bool

func (f messageFilter) toFilter() eventual2go.Filter {
	return func(d eventual2go.Data) bool { return f(d.(message)) }
}

func tomessageFilterArray(f ...messageFilter) (filter []eventual2go.Filter){

	filter = make([]eventual2go.Filter, len(f))
	for i, el := range f {
		filter[i] = el.toFilter()
	}
	return
}

func (s *messageStream) Where(f ...messageFilter) *messageStream {
	return &messageStream{s.Stream.Where(tomessageFilterArray(f...)...)}
}

func (s *messageStream) WhereNot(f ...messageFilter) *messageStream {
	return &messageStream{s.Stream.WhereNot(tomessageFilterArray(f...)...)}
}

func (s *messageStream) TransformWhere(t eventual2go.Transformer, f ...messageFilter) *eventual2go.Stream {
	return s.Stream.TransformWhere(t, tomessageFilterArray(f...)...)
}

func (s *messageStream) Split(f messageFilter) (*messageStream, *messageStream)  {
	return s.Where(f), s.WhereNot(f)
}

func (s *messageStream) First() *messageFuture {
	return &messageFuture{s.Stream.First()}
}

func (s *messageStream) FirstWhere(f... messageFilter) *messageFuture {
	return &messageFuture{s.Stream.FirstWhere(tomessageFilterArray(f...)...)}
}

func (s *messageStream) FirstWhereNot(f ...messageFilter) *messageFuture {
	return &messageFuture{s.Stream.FirstWhereNot(tomessageFilterArray(f...)...)}
}

func (s *messageStream) AsChan() (c chan message, stop *eventual2go.Completer) {
	c = make(chan message)
	stop = s.Listen(pipeTomessageChan(c))
	stop.Future().Then(closemessageChan(c))
	return
}

func pipeTomessageChan(c chan message) messageSubscriber {
	return func(d message) {
		c <- d
	}
}

func closemessageChan(c chan message) eventual2go.CompletionHandler {
	return func(d eventual2go.Data) eventual2go.Data {
		close(c)
		return nil
	}
}

type messageCollector struct {
	*eventual2go.Collector
}

func NewmessageCollector() *messageCollector {
	return &messageCollector{eventual2go.NewCollector()}
}

func (c *messageCollector) Add(d message) {
	c.Collector.Add(d)
}

func (c *messageCollector) AddFuture(f *messageFuture) {
	c.Collector.Add(f.Future)
}

func (c *messageCollector) AddStream(s *messageStream) {
	c.Collector.AddStream(s.Stream)
}

func (c *messageCollector) Get() message {
	return c.Collector.Get().(message)
}

func (c *messageCollector) Preview() message {
	return c.Collector.Preview().(message)
}

type messageObservable struct {
	*eventual2go.Observable
}

func NewmessageObservable (value message) (o *messageObservable) {
	return &messageObservable{eventual2go.NewObservable(value)}
}

func (o *messageObservable) Value() message {
	return o.Observable.Value().(message)
}

func (o *messageObservable) Change(value message) {
	o.Observable.Change(value)
}

func (o *messageObservable) OnChange(s messageSubscriber) (cancel *eventual2go.Completer) {
	return o.Observable.OnChange(s.toSubscriber())
}

func (o *messageObservable) Stream() (*messageStream) {
	return &messageStream{o.Observable.Stream()}
}


func (o *messageObservable) AsChan() (c chan message, cancel *eventual2go.Completer) {
	return o.Stream().AsChan()
}

func (o *messageObservable) NextChange() (f *messageFuture) {
	return o.Stream().First()
}

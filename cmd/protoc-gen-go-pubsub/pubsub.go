package main

import (
	"google.golang.org/protobuf/compiler/protogen"
)

const (
	contextPackage = protogen.GoImportPath("context")
	pubsubPackage  = protogen.GoImportPath("github.com/quarks-tech/pubsub-go/pubsub")
)

func generateFile(gen *protogen.Plugin, f *protogen.File) {
	filename := f.GeneratedFilenamePrefix + "_pubsub.pb.go"

	g := gen.NewGeneratedFile(filename, f.GoImportPath)
	g.P("// Code generated by protoc-gen-go-pubsub. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-go-pubsub v", version)
	g.P("// - protoc               v", protocVersion(gen))

	if f.Proto.GetOptions().GetDeprecated() {
		g.P("// ", f.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", f.Desc.Path())
	}

	g.P()
	g.P("package ", f.GoPackageName)
	g.P()

	generateFileContent(f, g)
}

func generateFileContent(f *protogen.File, g *protogen.GeneratedFile) {
	genPubSub(g, f, filterEventDataMessages(f.Messages))
}

func genPubSub(g *protogen.GeneratedFile, f *protogen.File, messages []*protogen.Message) {
	publisherName := "EventPublisher"
	subscriberName := "EventSubscriber"

	// Publisher interface
	g.P("type ", publisherName, " interface {")
	for _, m := range messages {
		g.P("Publish", removeMessageSuffix(m.GoIdent.GoName), "Event", "(ctx ", contextPackage.Ident("Context"), ", data *", m.GoIdent.GoName, ", opts ...", pubsubPackage.Ident("PublishOption"), ") error")
	}
	g.P("}")
	g.P()

	// Publisher struct
	g.P("type ", unexport(publisherName), " struct {")
	g.P("pp ", pubsubPackage.Ident("Publisher"))
	g.P("}")
	g.P()

	// Publisher factory
	g.P("func New", publisherName, " (pp ", pubsubPackage.Ident("Publisher"), ") ", publisherName, " {")
	g.P("return &", unexport(publisherName), "{pp}")
	g.P("}")
	g.P()

	// Publisher implementation
	for _, m := range messages {
		eventName := removeMessageSuffix(m.GoIdent.GoName)

		g.P("func (p *", unexport(publisherName), ")", " Publish", eventName, "Event", "(ctx ", contextPackage.Ident("Context"), ", data *", m.GoIdent.GoName, ", opts ...", pubsubPackage.Ident("PublishOption"), ") error {")
		g.P("return p.pp.Publish(ctx, ", quote(removeMessageSuffix(string(m.Desc.FullName()))), ", data, opts...)")
		g.P("}")
		g.P()
	}

	// Subscriber interface
	for _, m := range messages {
		eventName := removeMessageSuffix(m.GoIdent.GoName)

		g.P("type ", eventName, subscriberName, " interface {")
		g.P("On", eventName, "(ctx ", contextPackage.Ident("Context"), ", data *", m.GoIdent.GoName, ") error")
		g.P("}")
		g.P()
	}

	// Subscriber registration
	for _, m := range messages {
		eventName := removeMessageSuffix(m.GoIdent.GoName)

		g.P("func Register", eventName, subscriberName, "(r ", pubsubPackage.Ident("SubscriberRegistrar"), ", s ", eventName, subscriberName, ") {")
		g.P("r.RegisterSubscriber(&", eventName, "SubscriberDesc", ", s)")
		g.P("}")
		g.P()
	}

	// Handler
	for _, m := range messages {
		eventName := removeMessageSuffix(m.GoIdent.GoName)

		g.P("func ", eventName, "Handler(s interface{}, ctx ", contextPackage.Ident("Context"), ", dec func(interface{}) error, interceptor ", pubsubPackage.Ident("SubscriberInterceptor"), ") error {")
		g.P("data := new(", m.GoIdent.GoName, ")")
		g.P("if err := dec(data); err != nil {")
		g.P("return err")
		g.P("}")
		g.P("if interceptor == nil {")
		g.P("return s.(", eventName, subscriberName, ").", " On", eventName, "(ctx, data)")
		g.P("}")
		g.P("info := &", pubsubPackage.Ident("SubscriberInfo"), "{")
		g.P("Subscriber: s,")
		g.P("EventName: ", quote(removeMessageSuffix(string(m.Desc.FullName()))), ",")
		g.P("}")
		g.P("handler := func(ctx context.Context, data interface{}) error {")
		g.P("return s.(", eventName, subscriberName, ").", " On", eventName, "(ctx, data.(*", m.GoIdent.GoName, "))")
		g.P("}")
		g.P("return interceptor(ctx, data, info, handler)")
		g.P("}")
		g.P()
	}

	// Subscriber description
	g.P("var (")
	for _, m := range messages {
		eventName := removeMessageSuffix(m.GoIdent.GoName)

		g.P(eventName, "SubscriberDesc = ", pubsubPackage.Ident("SubscriberDesc"), "{")
		g.P("EventName: ", quote(removeMessageSuffix(string(m.Desc.FullName()))), ",")
		g.P("HandlerType: (*", eventName, subscriberName, ")(nil),")
		g.P("Handler: ", eventName, "Handler,")
		g.P("}")
	}
	g.P(")")
}
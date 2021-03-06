package inspector_test

import (
	"testing"

	"github.com/apoydence/onpar"
	. "github.com/apoydence/onpar/expect"
	. "github.com/apoydence/onpar/matchers"
	"github.com/apoydence/pubsub/pubsub-gen/internal/inspector"
)

type TL struct {
	*testing.T
	l inspector.Linker
}

func TestLinker(t *testing.T) {
	t.Parallel()
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) TL {
		return TL{
			T: t,
			l: inspector.NewLinker(),
		}
	})

	o.Spec("moves any known field types to PeerTypeFields", func(t TL) {
		m := map[string]inspector.Struct{
			"X": {Fields: []inspector.Field{
				{Name: "A", Type: "string"},
				{Name: "B", Type: "Y"},
				{Name: "C", Type: "Y"},
			}},
			"Y": {Fields: []inspector.Field{
				{Name: "A", Type: "string"},
				{Name: "B", Type: "int"},
			}},
		}
		t.l.Link(m, nil)

		Expect(t, m["X"].Fields).To(HaveLen(1))
		Expect(t, m["X"].Fields[0].Name).To(Equal("A"))

		Expect(t, m["X"].PeerTypeFields).To(HaveLen(2))
		Expect(t, m["X"].PeerTypeFields[0].Name).To(Equal("B"))
		Expect(t, m["X"].PeerTypeFields[1].Name).To(Equal("C"))
	})

	o.Spec("moves any known interface types to InterfaceTypeFields", func(t TL) {
		m := map[string]inspector.Struct{
			"X": {Fields: []inspector.Field{
				{Name: "A", Type: "string"},
				{Name: "B", Type: "Y"},
				{Name: "C", Type: "Y"},
			}},
			"Y": {Fields: []inspector.Field{
				{Name: "A", Type: "string"},
				{Name: "B", Type: "int"},
				{Name: "C", Type: "MyInterfaceThing"},
			}},
		}

		mi := map[string][]string{
			"MyInterfaceThing": {
				"X", "Y",
			},
		}
		c := m["Y"].Fields[2]

		t.l.Link(m, mi)
		Expect(t, m["Y"].Fields).To(HaveLen(2))
		Expect(t, m["Y"].InterfaceTypeFields).To(HaveLen(1))
		Expect(t, m["Y"].InterfaceTypeFields[c]).To(And(
			HaveLen(2),
			Contain("X", "Y"),
		))
	})
}

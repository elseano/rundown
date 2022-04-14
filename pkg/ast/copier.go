package ast

import goldast "github.com/yuin/goldmark/ast"

func CopyNode(node goldast.Node) goldast.Node {
	switch n := node.(type) {

	case *goldast.Paragraph:
		new := goldast.NewParagraph()

		CopySettings(n, new)
		CopyChildren(n, new)

		return new

	case *goldast.Text:
		return goldast.NewTextSegment(n.Segment)

	case *goldast.Heading:
		new := goldast.NewHeading(n.Level)

		CopySettings(n, new)
		CopyChildren(n, new)

		return new

	case *goldast.String:
		new := goldast.NewString(n.Value)

		CopySettings(n, new)

		return new

	case *ExecutionBlock:
		new := NewExecutionBlock(n.CodeBlock)

		CopySettings(n, new)

		new.ifScript = n.ifScript
		new.CaptureEnvironment = n.CaptureEnvironment
		new.CaptureStdoutInto = n.CaptureStdoutInto
		new.Execute = n.Execute
		new.Language = n.Language
		new.ReplaceProcess = n.ReplaceProcess
		new.Reveal = n.Reveal
		new.ShowStderr = n.ShowStderr
		new.ShowStdout = n.ShowStdout
		new.SkipOnFailure = n.SkipOnFailure
		new.SkipOnSuccess = n.SkipOnSuccess
		new.SpinnerMode = n.SpinnerMode
		new.SpinnerName = n.SpinnerName
		new.SubstituteEnvironment = n.SubstituteEnvironment
		new.With = n.With

		return new

	case *goldast.FencedCodeBlock:
		new := goldast.NewFencedCodeBlock(n.Info)

		CopySettings(n, new)

		return new

	case *goldast.Emphasis:
		new := goldast.NewEmphasis(n.Level)
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *goldast.CodeBlock:
		new := goldast.NewCodeBlock()
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *DescriptionBlock:
		new := NewDescriptionBlock()
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *SaveCodeBlock:
		new := NewSaveCodeBlock(n.CodeBlock, n.SaveToVariable)
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *StopFail:
		new := NewStopFail()
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *StopOk:
		new := NewStopOk()
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *SubEnvBlock:
		new := NewSubEnvBlock(n.InnerBlock)
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *EnvironmentSubstitution:
		new := NewSubEnvInline(string(n.Value))
		CopySettings(n, new)
		CopyChildren(n, new)
		return new

	case *InvokeBlock:
		new := NewInvokeBlock()
		new.Args = n.Args
		new.AsDependency = n.AsDependency
		new.Invoke = n.Invoke
		CopySettings(n, new)
		CopyChildren(n, new)

		return new

	}

	return nil
}

func CopyChildren(from goldast.Node, to goldast.Node) {
	for child := from.FirstChild(); child != nil; child = child.NextSibling() {
		copied := CopyNode(child)
		if copied != nil {
			to.AppendChild(to, copied)
		}
	}
}

func CopySettings(from goldast.Node, to goldast.Node) {
	type inline interface {
		Inline()
	}

	if _, isInline := to.(inline); !isInline {
		to.SetLines(from.Lines())
	}

	if fromIf, ok := from.(Conditional); ok {
		toIf := to.(Conditional)

		toIf.SetIfScript(fromIf.GetIfScript())
	}
}

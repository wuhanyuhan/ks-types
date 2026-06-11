package kstypes

import "testing"

func TestSDUIPrimitiveSchemas_Coverage(t *testing.T) {
	// 首批原语必须全部注册
	want := []string{
		PrimitiveStack, PrimitiveGrid, PrimitiveCard, PrimitiveSection, PrimitiveTabs, PrimitiveSplit,
		PrimitiveText, PrimitiveMarkdown, PrimitiveFieldGroup, PrimitiveTable, PrimitiveStatusBadge, PrimitiveMetric, PrimitiveEmptyState,
		PrimitiveButton, PrimitiveForm, PrimitiveLink,
		PrimitiveListActions, PrimitiveDiffReview, PrimitiveTimeline, PrimitiveCardGrid, PrimitiveImageVariants,
	}
	for _, k := range want {
		if _, ok := SDUIPrimitiveSchemas[k]; !ok {
			t.Errorf("primitive %q not registered in SDUIPrimitiveSchemas", k)
		}
	}
}

func TestContainerPrimitives_OnlyContainers(t *testing.T) {
	for _, c := range []string{PrimitiveStack, PrimitiveGrid, PrimitiveCard, PrimitiveSection, PrimitiveTabs, PrimitiveSplit} {
		if !ContainerPrimitives[c] {
			t.Errorf("container primitive %q missing from ContainerPrimitives", c)
		}
	}
	for _, leaf := range []string{PrimitiveText, PrimitiveButton, PrimitiveTable, PrimitiveListActions} {
		if ContainerPrimitives[leaf] {
			t.Errorf("leaf primitive %q wrongly marked as container", leaf)
		}
	}
}

func TestSDUIStackProps_Validate(t *testing.T) {
	if err := (SDUIStackProps{Direction: "vertical"}).Validate(); err != nil {
		t.Errorf("valid stack rejected: %v", err)
	}
	if err := (SDUIStackProps{Direction: "diagonal"}).Validate(); err == nil {
		t.Error("invalid direction accepted")
	}
}

func TestSDUIGridProps_Validate(t *testing.T) {
	if err := (SDUIGridProps{Columns: 3}).Validate(); err != nil {
		t.Errorf("valid grid rejected: %v", err)
	}
	if err := (SDUIGridProps{Columns: 9}).Validate(); err == nil {
		t.Error("grid.columns=9 accepted")
	}
}

func TestSDUIButtonProps_Validate(t *testing.T) {
	if err := (SDUIButtonProps{Label: "确认", Action: SDUIActionIntent{ActionID: "approve"}}).Validate(); err != nil {
		t.Errorf("valid button rejected: %v", err)
	}
	if err := (SDUIButtonProps{Label: "", Action: SDUIActionIntent{ActionID: "x"}}).Validate(); err == nil {
		t.Error("empty label accepted")
	}
	if err := (SDUIButtonProps{Label: "确认", Action: SDUIActionIntent{}}).Validate(); err == nil {
		t.Error("empty action_id accepted")
	}
}

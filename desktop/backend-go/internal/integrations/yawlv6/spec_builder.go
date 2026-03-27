package yawlv6

import (
	"fmt"
	"strings"
)

const yawlNS = `xmlns="http://www.citi.qut.edu.au/yawl" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:schemaLocation="http://www.citi.qut.edu.au/yawl YAWL_Schema.xsd"`

// BuildSequenceSpec generates a YAWL WCP-1 (Sequence) specification XML for
// a linear chain of tasks.  Returns an empty string if tasks is empty.
func BuildSequenceSpec(tasks []string) string {
	if len(tasks) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><specificationSet %s><specification uri="OSA_Sequence"><metaData/><rootNet id="Net"><processControlElements>`,
		yawlNS,
	))
	// InputCondition flows into first task.
	sb.WriteString(fmt.Sprintf(
		`<inputCondition id="InputCondition"><flowsInto><nextElementRef id="%s"/></flowsInto></inputCondition>`,
		tasks[0],
	))
	// Chain tasks; last task flows into OutputCondition.
	for i, task := range tasks {
		next := "OutputCondition"
		if i < len(tasks)-1 {
			next = tasks[i+1]
		}
		sb.WriteString(fmt.Sprintf(
			`<task id="%s"><flowsInto><nextElementRef id="%s"/></flowsInto><join code="xor"/><split code="and"/></task>`,
			task, next,
		))
	}
	sb.WriteString(`<outputCondition id="OutputCondition"/></processControlElements></rootNet></specification></specificationSet>`)
	return sb.String()
}

// BuildParallelSplitSpec generates a YAWL WCP-2 (Parallel Split) specification XML.
// The trigger task fans out with an AND-split into all branch tasks, each of which
// then flows directly to the OutputCondition.
func BuildParallelSplitSpec(trigger string, branches []string) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf(
		`<?xml version="1.0" encoding="UTF-8"?><specificationSet %s><specification uri="OSA_ParallelSplit"><metaData/><rootNet id="Net"><processControlElements>`,
		yawlNS,
	))
	sb.WriteString(fmt.Sprintf(
		`<inputCondition id="InputCondition"><flowsInto><nextElementRef id="%s"/></flowsInto></inputCondition>`,
		trigger,
	))
	// Trigger task: AND-split flowing into every branch.
	var triggerFlow strings.Builder
	for _, b := range branches {
		triggerFlow.WriteString(fmt.Sprintf(`<flowsInto><nextElementRef id="%s"/></flowsInto>`, b))
	}
	sb.WriteString(fmt.Sprintf(
		`<task id="%s">%s<join code="xor"/><split code="and"/></task>`,
		trigger, triggerFlow.String(),
	))
	// Branch tasks flow into OutputCondition.
	for _, b := range branches {
		sb.WriteString(fmt.Sprintf(
			`<task id="%s"><flowsInto><nextElementRef id="OutputCondition"/></flowsInto><join code="xor"/><split code="and"/></task>`,
			b,
		))
	}
	sb.WriteString(`<outputCondition id="OutputCondition"/></processControlElements></rootNet></specification></specificationSet>`)
	return sb.String()
}

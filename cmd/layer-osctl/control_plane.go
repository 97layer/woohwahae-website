package main

import (
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"layer-os/internal/runtime"
)

type controlPlaneService interface {
	authStatusProvider
	ListPreflights() []runtime.PreflightRecord
	CreatePreflight(item runtime.PreflightRecord) error
	ListVerifications() []runtime.VerificationRecord
	RunVerification(recordID string, scope string, workdir string, command []string, baseNotes []string) (runtime.VerificationRecord, error)
	ListPolicies() []runtime.PolicyDecision
	EvaluatePolicy(decisionID string, intent string, scope string, risk string, novelty string, tokenClass string, requiresApproval bool) (runtime.PolicyDecision, error)
	ListGatewayCalls() []runtime.GatewayCall
	CreateGatewayCall(item runtime.GatewayCall) error
	ListEvents() []runtime.EventEnvelope
	CreateEvent(input runtime.EventCreateInput) (runtime.EventEnvelope, error)
	StreamEvents(out io.Writer) error
	ListExecutes() []runtime.ExecuteRun
	RunExecute(executeID string, workItemID string, policyDecisionID string, baseNotes []string) (runtime.ExecuteRun, error)
}

func runPreflight(service controlPlaneService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl preflight <list|create>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListPreflights())
	case "create":
		createPreflight(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl preflight <list|create>")
	}
}

func runVerify(service controlPlaneService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl verify <list|run>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListVerifications())
	case "run":
		runVerification(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl verify <list|run>")
	}
}

func runPolicy(service controlPlaneService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl policy <list|evaluate>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListPolicies())
	case "evaluate":
		evaluatePolicy(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl policy <list|evaluate>")
	}
}

func runGateway(service controlPlaneService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl gateway <list|create>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListGatewayCalls())
	case "create":
		createGateway(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl gateway <list|create>")
	}
}

func runEvent(service controlPlaneService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl event <list|create|stream>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListEvents())
	case "create":
		createEvent(service, args[1:])
	case "stream":
		streamEvents(service)
	default:
		log.Fatal("usage: layer-osctl event <list|create|stream>")
	}
}

func runExecute(service controlPlaneService, args []string) {
	if len(args) == 0 {
		log.Fatal("usage: layer-osctl execute <list|run>")
	}
	switch args[0] {
	case "list":
		writeJSON(service.ListExecutes())
	case "run":
		runExecuteRecord(service, args[1:])
	default:
		log.Fatal("usage: layer-osctl execute <list|run>")
	}
}

func createPreflight(service controlPlaneService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("preflight create", flag.ExitOnError)
	id := cmd.String("id", "", "preflight record id")
	task := cmd.String("task", "", "task summary")
	mode := cmd.String("mode", "internal", "preflight mode")
	status := cmd.String("status", "ready", "preflight status")
	decision := cmd.String("decision", "go", "preflight decision")
	models := cmd.String("models", "", "comma-separated model ids")
	steps := cmd.String("steps", "", "comma-separated plan steps")
	risks := cmd.String("risks", "", "comma-separated risks")
	checks := cmd.String("checks", "", "comma-separated checks")
	parseArgs(cmd, args)

	item := runtime.PreflightRecord{
		RecordID:    strings.TrimSpace(*id),
		Task:        strings.TrimSpace(*task),
		Mode:        strings.TrimSpace(*mode),
		Status:      strings.TrimSpace(*status),
		Decision:    strings.TrimSpace(*decision),
		ModelsUsed:  mergeModelLists(splitCSV(*models), requestModels()),
		Steps:       splitCSV(*steps),
		Risks:       splitCSV(*risks),
		Checks:      splitCSV(*checks),
		GeneratedAt: time.Now().UTC(),
	}
	if err := service.CreatePreflight(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func evaluatePolicy(service controlPlaneService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("policy evaluate", flag.ExitOnError)
	id := cmd.String("id", "", "policy decision id")
	intent := cmd.String("intent", "", "policy intent")
	scope := cmd.String("scope", "kernel", "policy scope")
	risk := cmd.String("risk", "low", "policy risk")
	novelty := cmd.String("novelty", "low", "policy novelty")
	tokenClass := cmd.String("token-class", "small", "policy token class")
	requiresApproval := cmd.Bool("requires-approval", false, "requires approval")
	parseArgs(cmd, args)

	item, err := service.EvaluatePolicy(
		strings.TrimSpace(*id),
		strings.TrimSpace(*intent),
		strings.TrimSpace(*scope),
		strings.TrimSpace(*risk),
		strings.TrimSpace(*novelty),
		strings.TrimSpace(*tokenClass),
		*requiresApproval,
	)
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func createGateway(service controlPlaneService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("gateway create", flag.ExitOnError)
	id := cmd.String("id", "", "gateway call id")
	decisionID := cmd.String("decision", "", "policy decision id")
	provider := cmd.String("provider", "", "gateway provider")
	model := cmd.String("model", "", "gateway model")
	kind := cmd.String("kind", "", "gateway request kind")
	status := cmd.String("status", "recorded", "gateway status")
	budget := cmd.Int("budget", 0, "gateway token budget")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item := runtime.GatewayCall{
		CallID:      strings.TrimSpace(*id),
		DecisionID:  strings.TrimSpace(*decisionID),
		Provider:    strings.TrimSpace(*provider),
		Model:       strings.TrimSpace(*model),
		RequestKind: strings.TrimSpace(*kind),
		Status:      strings.TrimSpace(*status),
		TokenBudget: *budget,
		Notes:       splitCSV(*notes),
		CreatedAt:   time.Now().UTC(),
	}
	if err := service.CreateGatewayCall(item); err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func runExecuteRecord(service controlPlaneService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("execute run", flag.ExitOnError)
	id := cmd.String("id", "", "execute run id")
	workID := cmd.String("work", "", "work item id")
	decisionID := cmd.String("decision", "", "policy decision id")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	item, err := service.RunExecute(strings.TrimSpace(*id), strings.TrimSpace(*workID), strings.TrimSpace(*decisionID), splitCSV(*notes))
	if item.ExecuteID != "" {
		writeJSON(item)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func runVerification(service controlPlaneService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("verify run", flag.ExitOnError)
	id := cmd.String("id", "", "verification record id")
	scope := cmd.String("scope", "kernel", "verification scope")
	notes := cmd.String("notes", "", "comma-separated notes")
	parseArgs(cmd, args)

	root := strings.TrimSpace(os.Getenv("LAYER_OS_REPO_ROOT"))
	if root == "" {
		root = "."
	}
	command, err := runtime.DefaultGoTestVerificationCommand()
	if err != nil {
		log.Fatal(err)
	}
	item, err := service.RunVerification(strings.TrimSpace(*id), strings.TrimSpace(*scope), root, command, splitCSV(*notes))
	if item.RecordID != "" {
		writeJSON(item)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func createEvent(service controlPlaneService, args []string) {
	requireCLIWriteAuth(service)
	cmd := flag.NewFlagSet("event create", flag.ExitOnError)
	kind := cmd.String("kind", "", "event kind")
	surface := cmd.String("surface", string(runtime.SurfaceAPI), "event surface")
	workID := cmd.String("work", "system", "work item id")
	stage := cmd.String("stage", string(runtime.StageDiscover), "event stage")
	data := cmd.String("data", "", "comma-separated key=value pairs")
	parseArgs(cmd, args)

	item, err := service.CreateEvent(runtime.EventCreateInput{
		Kind:       strings.TrimSpace(*kind),
		Surface:    runtime.Surface(strings.TrimSpace(*surface)),
		WorkItemID: strings.TrimSpace(*workID),
		Stage:      runtime.Stage(strings.TrimSpace(*stage)),
		Data:       splitPairs(*data),
	})
	if err != nil {
		log.Fatal(err)
	}
	writeJSON(item)
}

func streamEvents(service controlPlaneService) {
	if err := service.StreamEvents(os.Stdout); err != nil {
		log.Fatal(err)
	}
}

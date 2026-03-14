package main

import "log"

type rootCommandHandler func(cliService, []string)

var rootCommandRegistry = map[string]rootCommandHandler{
	"status": func(service cliService, args []string) {
		runStatus(service, args)
	},
	"daemon": func(service cliService, args []string) {
		daemonService, ok := service.(daemonControlService)
		if !ok {
			log.Fatal("daemon status unsupported")
		}
		runDaemonCommand(daemonService, args)
	},
	"handoff": func(service cliService, args []string) {
		runHandoff(service, args)
	},
	"knowledge": func(service cliService, args []string) {
		writeJSON(service.Knowledge())
	},
	"next": func(service cliService, args []string) {
		nextService, ok := service.(nextService)
		if !ok {
			log.Fatal("next unsupported")
		}
		runNext(nextService, args)
	},
	"telegram": func(service cliService, args []string) {
		runTelegram(service, args)
	},
	"threads": func(service cliService, args []string) {
		runThreads(service, args)
	},
	"session": func(service cliService, args []string) {
		runSession(service, args)
	},
	"review-room": func(service cliService, args []string) {
		runReviewRoom(service, args)
	},
	"writer": func(service cliService, args []string) {
		writeJSON(service.WriteLease())
	},
	"adapters": func(service cliService, args []string) {
		writeJSON(service.Adapters())
	},
	"provider": func(service cliService, args []string) {
		runProvider(service, args)
	},
	"ingest": func(service cliService, args []string) {
		runIngest(service, args)
	},
	"corpus": func(service cliService, args []string) {
		runCorpus(service, args)
	},
	"conversation": func(service cliService, args []string) {
		runConversation(service, args)
	},
	"observation": func(service cliService, args []string) {
		runObservation(service, args)
	},
	"thread": func(service cliService, args []string) {
		runThread(service, args)
	},
	"capabilities": func(service cliService, args []string) {
		writeJSON(service.Capabilities())
	},
	"auth": func(service cliService, args []string) {
		runAuth(service, args)
	},
	"preflight": func(service cliService, args []string) {
		runPreflight(service, args)
	},
	"policy": func(service cliService, args []string) {
		runPolicy(service, args)
	},
	"gateway": func(service cliService, args []string) {
		runGateway(service, args)
	},
	"event": func(service cliService, args []string) {
		runEvent(service, args)
	},
	"execute": func(service cliService, args []string) {
		runExecute(service, args)
	},
	"audit": func(service cliService, args []string) {
		runAudit(args)
	},
	"verify": func(service cliService, args []string) {
		runVerify(service, args)
	},
	"branch": func(service cliService, args []string) {
		runBranch(service, args)
	},
	"proposal": func(service cliService, args []string) {
		runProposal(service, args)
	},
	"job": func(service cliService, args []string) {
		runJob(service, args)
	},
	"quickwork": func(service cliService, args []string) {
		quickWorkService, ok := service.(quickWorkService)
		if !ok {
			log.Fatal("quickwork unsupported")
		}
		runQuickWork(quickWorkService, args)
	},
	"work": func(service cliService, args []string) {
		runWork(service, args)
	},
	"founder": func(service cliService, args []string) {
		runFounder(service, args)
	},
	"flow": func(service cliService, args []string) {
		runFlow(service, args)
	},
	"approval": func(service cliService, args []string) {
		runApproval(service, args)
	},
	"release": func(service cliService, args []string) {
		runRelease(service, args)
	},
	"deploy": func(service cliService, args []string) {
		runDeploy(service, args)
	},
	"rollback": func(service cliService, args []string) {
		runRollback(service, args)
	},
	"memory": func(service cliService, args []string) {
		runMemory(service, args)
	},
	"target": func(service cliService, args []string) {
		runTarget(service, args)
	},
	"snapshot": func(service cliService, args []string) {
		runSnapshot(service, args)
	},
	"smoke": func(service cliService, args []string) {
		runSmoke(args)
	},
}

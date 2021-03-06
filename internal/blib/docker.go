package blib

import (
	"bufio"
	"fmt"
	"github.com/willf/pad"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

var (
	composeError	error
	dockerError		error
)


func NewDocker() bool {

	if Fd.FdBuildContext == "docker" {
		DeploymentHead()
		PipelineHead()
	}

	return dockerDeploy(dockerBuild())
}


func dockerBuild() bool {
	if Fd.FdClean { dockerClean() }

	success := composePush(composeBuild())

	return success
}


func dockerDeploy(prevSuccess bool) bool {
	success := false
	if prevSuccess {
		if Fd.FdLocal {
			success = composeUp(composeRemove(composeStop(composePull())))
		} else {
			if success = NewGcp();	!success { dockerError = GetGcpError() }
		}
	}

	return success
}


func composeBuild() bool {
	logPrefix	:= Yellow(pad.Right("\ncomposeBuild():", 20, " "))
	args		:= "build --no-cache --pull "	+		// Add future build args here.
					"--build-arg NICKNAME "		+		// BTW, an absent value here(as opposed to "--build-arg NICKNAME=foo")
					"--build-arg SERVICE_NAME " + 		// forces the build to reference the calling ENV for the value
					"--build-arg TARGET_ALIAS " +
						Fd.FdServiceName
	argsAbbrev	:= "build (...) " + Fd.FdServiceName

	return composeRun(logPrefix, args, argsAbbrev)
}


func composePush(prevSuccess bool) bool {
	if !prevSuccess{ return false }

	logPrefix	:= Yellow(pad.Right("\ncomposePush():", 20, " "))
	args		:= "push " + Fd.FdServiceName
	argsAbbrev	:= args

	return composeRun(logPrefix, args, argsAbbrev)
}


func composePull() bool {
	logPrefix	:= Yellow(pad.Right("\ncomposePull():", 20, " "))
	args		:= "pull " + Fd.FdServiceName
	argsAbbrev	:= args

	return composeRun(logPrefix, args, argsAbbrev)
}


func composeUp(prevSuccess bool) bool {
	if !prevSuccess { return false }

	logPrefix	:= Yellow(pad.Right("\ncomposeUp():", 20, " "))
	args		:= "--log-level " + Fd.FdTargetLogLevel + " up --detach --force-recreate " + Fd.FdServiceName
	argsAbbrev	:= "(...) up (...) " + Fd.FdServiceName

	return composeRun(logPrefix, args, argsAbbrev)
}


func composeStop(prevSuccess bool) bool {
	if !prevSuccess { return false }

	logPrefix	:= Yellow(pad.Right("\ncomposeStop():", 20, " "))
	args		:= "stop " + Fd.FdServiceName
	argsAbbrev	:= args

	return composeRun(logPrefix, args, argsAbbrev)
}


func composeRemove(prevSuccess bool) bool {
	if !prevSuccess { return false }

	logPrefix	:= Yellow(pad.Right("\ncomposeRemove():", 20, " "))
	args		:= "rm --force " + Fd.FdServiceName
	argsAbbrev	:= args

	return composeRun(logPrefix, args, argsAbbrev)
}


func composeRun(prefix string, cmdArgs string, cmdArgsAbbrev string) bool {
	if Fd.FdVerbose {
		logCommand	:= BlackOnGray(" docker-compose " + cmdArgs + " ")
		fmt.Printf("%s$ %s", prefix, logCommand)
		fmt.Printf("\n")
	} else {
		logCommand	:= "docker-compose " + cmdArgsAbbrev
		fmt.Printf("%s$ %s", prefix, logCommand)
	}

	command			:=	exec.Command("docker-compose", strings.Split(cmdArgs, " ")...)
	if success		:=	setEnvironment(); !success { log.Println("Curious issue with setting the environment :'(")}
	command.Env		=	os.Environ()
	if Fd.FdDebug {
		fmt.Printf("\n")
		log.Println("DEBUG:",				os.Getenv("DEBUG"))
		log.Println("LOGS:",				os.Getenv("LOGS"))
		log.Println("NICKNAME:",			os.Getenv("NICKNAME"))
		log.Println("SERVICE_NAME:",		os.Getenv("SERVICE_NAME"))
		log.Println("SITE_NICKNAME:",		os.Getenv("SITE_NICKNAME"))
		log.Println("TARGET_ALIAS:",		os.Getenv("TARGET_ALIAS"))
		log.Println("TARGET_IMAGE_TAG:",	os.Getenv("TARGET_IMAGE_TAG"))
		log.Println("TARGET_LOCAL_PORT:",	os.Getenv("TARGET_LOCAL_PORT"))
		log.Println("TARGET_LOG_LEVEL:",	os.Getenv("TARGET_LOG_LEVEL"))
		log.Println("TARGET_PROJECT_ID:",	os.Getenv("TARGET_PROJECT_ID"))
		log.Println("TARGET_REMOTE_PORT:",	os.Getenv("TARGET_REMOTE_PORT"))
	}

	stderr, _		:= command.StderrPipe()
	composeError	= command.Start()
	if composeError != nil { log.Printf("%s", Red(composeError)) }

	scanner			:= bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		stderrText := scanner.Text()
		if Fd.FdVerbose { log.Printf("%s", Grey(stderrText)) }
	}

	composeError = command.Wait()
	if composeError != nil {
		log.Printf("%s$ %s%s", prefix, command, WhiteOnRed(" X "))
		log.Fatalf("\n%s", Red(composeError))
	}

	return true
}


func dockerClean() bool {
	logPrefix	:= Yellow(pad.Right("\ndockerClean():", 20, " "))

	args		:= "container prune --force"
	argsAbbrev	:= args
	dockerRun(logPrefix, args, argsAbbrev)

	args		= "image prune --all --force"
	argsAbbrev	= args
	dockerRun(logPrefix, args, argsAbbrev)

	args		= "network prune --force"
	argsAbbrev	= args
	dockerRun(logPrefix, args, argsAbbrev)

	args		= "volume prune --force"
	argsAbbrev	= args
	dockerRun(logPrefix, args, argsAbbrev)

	return true
}


func dockerRun(prefix string, cmdArgs string, cmdArgsAbbrev string) bool {
	if Fd.FdVerbose {
		logCommand	:= BlackOnGray(" docker " + cmdArgs + " ")
		fmt.Printf("%s$ %s", prefix, logCommand)
		fmt.Printf("\n")
	} else {
		logCommand	:= "docker " + cmdArgsAbbrev
		fmt.Printf("%s$ %s", prefix, logCommand)
	}

	command			:= exec.Command("docker", strings.Split(cmdArgs, " ")...)
	stderr, _		:= command.StderrPipe()
	dockerError		= command.Start(); if dockerError != nil { log.Printf("%s", Red(dockerError)) }
	scanner			:= bufio.NewScanner(stderr)

	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		stderrText := scanner.Text()
		if Fd.FdVerbose { log.Printf("%s", Grey(stderrText)) }
	}

	dockerError		= command.Wait()
	if dockerError != nil {
		log.Printf("%s$ %s%s", prefix, command, WhiteOnRed(" X "))
		log.Fatalf("\n%s", Red(dockerError))
	}

	return true
}


func setEnvironment() bool {
	if err := os.Setenv("DEBUG",				strconv.FormatBool(Fd.FdDebug)	); err != nil { println("derp"); return false }
	if err := os.Setenv("TEST",				strconv.FormatBool(Fd.FdTest)	); err != nil { println("derp"); return false }
	if err := os.Setenv("LOGS",				strconv.FormatBool(Fd.FdVerbose)); err != nil { println("derp"); return false }
	if err := os.Setenv("TARGET_LOCAL_PORT",	Fd.FdTargetLocalPort			); err != nil { println("derp"); return false }
	if err := os.Setenv("TARGET_LOG_LEVEL",	Fd.FdTargetLogLevel				); err != nil { println("derp"); return false }
	if err := os.Setenv("TARGET_REMOTE_PORT",	Fd.FdTargetRemotePort			); err != nil { println("derp"); return false }
	if err := os.Setenv("TARGET_PROJECT_ID",	Fd.FdTargetProjectId			); err != nil { println("derp"); return false }
	if err := os.Setenv("TARGET_ALIAS",		Fd.FdTargetAlias				); err != nil { println("derp"); return false }
	if err := os.Setenv("SERVICE_NAME",		Fd.FdServiceName				); err != nil { println("derp"); return false }
	if err := os.Setenv("TARGET_IMAGE_TAG",	Fd.FdTargetImageTag				); err != nil { println("derp"); return false }
	if err := os.Setenv("NICKNAME",			Fd.FdNickname					); err != nil { println("derp"); return false }
	if err := os.Setenv("SITE_NICKNAME",		Fd.FdSiteNickname				); err != nil { println("derp"); return false }

	return true
}


func GetComposeError()	error { return composeError	}

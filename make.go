package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// POSIX make spec:
// https://pubs.opengroup.org/onlinepubs/9699919799/utilities/make.html

type Target struct {
	sync.Once

	Prerequisites []string
	Commands      []string
}

func NewTarget(prerequisites []string) *Target {
	t := Target{
		Prerequisites: prerequisites,
	}
	return &t
}

type Graph map[string]*Target

func execute(graph Graph, name string) error {
	// lookup current target
	target, ok := graph[name]
	if !ok {
		return fmt.Errorf("target does not exist: %s", name)
	}

	errors := make(chan error)

	// execute any prerequisites (recursive call)
	var wg sync.WaitGroup
	for _, preprequisite := range target.Prerequisites {
		preprequisite := preprequisite

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := execute(graph, preprequisite)
			if err != nil {
				errors <- err
			}
		}()
	}

	// turn wg.Wait() into a select-able channel
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	// check for errors / wait for prerequisites to finish
	select {
	case <-done:
	case err := <-errors:
		return err
	}

	// execute current target (base case)
	var commandErr error
	target.Do(func() {
		for _, command := range target.Commands {
			fmt.Println(command)

			var env []string
			var prog string
			var args []string

			fields := strings.Fields(command)
			for _, field := range fields {
				// check for env vars
				if prog == "" && strings.Contains(field, "=") {
					env = append(env, field)
					continue
				}
				// set the prog name if not set
				if prog == "" {
					prog = field
					continue
				}
				// all other fields are args
				args = append(args, field)
			}

			cmd := exec.Command(prog, args...)
			cmd.Env = append(cmd.Environ(), env...)

			out, err := cmd.CombinedOutput()
			fmt.Print(string(out))
			if err != nil {
				commandErr = err
				break
			}
		}
	})

	return commandErr
}

func run() error {
	var fileName string
	flag.StringVar(&fileName, "f", "Makefile", "Read file as Makefile")
	flag.Parse()

	file, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		lines = append(lines, line)
	}

	err = scanner.Err()
	if err != nil {
		return err
	}

	graph := make(map[string]*Target)

	var current string
	for _, line := range lines {
		// ignore empty lines
		if len(line) == 0 {
			continue
		}
		// ignore comments
		if line[0] == '#' {
			continue
		}
		// ignore dot directives
		if line[0] == '.' {
			continue
		}

		// add commands to the current target
		if line[0] == '\t' {
			target, ok := graph[current]
			if !ok {
				return fmt.Errorf("target does not exist: %s", current)
			}

			command := strings.TrimSpace(line)
			target.Commands = append(target.Commands, command)

			continue
		}

		// create target and add prerequisites
		fields := strings.Fields(line)
		name, prerequisites := fields[0], fields[1:]
		name, _ = strings.CutSuffix(name, ":")
		graph[name] = NewTarget(prerequisites)

		// update current target
		current = name
	}

	name := "default"
	if len(os.Args) >= 2 {
		name = os.Args[1]
	}

	return execute(graph, name)
}

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(0)
}

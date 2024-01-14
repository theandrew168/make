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

// A target is a list of commands with possible dependencies
// (referenced via names). The `sync.Once` is used to ensure
// that each target is only executed a single time.
type Target struct {
	sync.Once

	Dependencies []string
	Commands     []string
}

func NewTarget(dependencies []string) *Target {
	t := Target{
		Dependencies: dependencies,
	}
	return &t
}

// The "graph" here is really just a mapping of names to targets.
type Graph map[string]*Target

// Read the lines of a text file.
func readLines(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
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
		return nil, err
	}

	return lines, nil
}

// Build a graph of targets from the lines of a Makefile.
func buildGraph(lines []string) (Graph, error) {
	graph := make(Graph)

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
				return nil, fmt.Errorf("target does not exist: %s", current)
			}

			command := strings.TrimSpace(line)
			target.Commands = append(target.Commands, command)

			continue
		}

		// create target and add dependencies
		fields := strings.Fields(line)
		name, dependencies := fields[0], fields[1:]
		name, _ = strings.CutSuffix(name, ":")
		graph[name] = NewTarget(dependencies)

		// update current target
		current = name
	}

	return graph, nil
}

// Execute the target after recursively executing any dependencies.
func execute(graph Graph, name string) error {
	// lookup current target
	target, ok := graph[name]
	if !ok {
		return fmt.Errorf("target does not exist: %s", name)
	}

	errors := make(chan error)

	// execute any dependencies (recursive call)
	var wg sync.WaitGroup
	for _, dependency := range target.Dependencies {
		dependency := dependency

		wg.Add(1)
		go func() {
			defer wg.Done()
			err := execute(graph, dependency)
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

	// check for errors / wait for dependencies to finish
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

	lines, err := readLines(fileName)
	if err != nil {
		return err
	}

	graph, err := buildGraph(lines)
	if err != nil {
		return err
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

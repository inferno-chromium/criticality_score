package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/ossf/criticality_score/cmd/collect_signals/collector"
	"github.com/ossf/criticality_score/cmd/collect_signals/github"
	"github.com/ossf/criticality_score/cmd/collect_signals/githubmentions"
	"github.com/ossf/criticality_score/cmd/collect_signals/result"
	"github.com/ossf/criticality_score/internal/githubapi"
	"github.com/ossf/criticality_score/internal/logflag"
	"github.com/ossf/criticality_score/internal/outfile"
	"github.com/ossf/scorecard/v4/clients/githubrepo/roundtripper"
	sclog "github.com/ossf/scorecard/v4/log"
	log "github.com/sirupsen/logrus"
	//csvutil (it's docs are heaps better!)
)

const defaultLogLevel = log.InfoLevel

var logFlag = logflag.Level(defaultLogLevel)

func init() {
	flag.Var(&logFlag, "log", "set the `level` of logging.")
	outfile.DefineFlags(flag.CommandLine, "force", "append", "OUT_FILE")
	flag.Usage = func() {
		cmdName := path.Base(os.Args[0])
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "Usage:\n  %s [FLAGS]... IN_FILE... OUT_FILE\n\n", cmdName)
		fmt.Fprintf(w, "Collects signals for each project repository listed.\n")
		fmt.Fprintf(w, "IN_FILE must be either a file or - to read from stdin.\n")
		fmt.Fprintf(w, "OUT_FILE must be either be a file or - to write to stdout.\n")
		fmt.Fprintf(w, "\nFlags:\n")
		flag.PrintDefaults()
	}
}

func main() {
	flag.Parse()

	logger := log.New()
	logger.SetLevel(logFlag.Level())

	// roundtripper requires us to use the scorecard logger.
	scLogger := sclog.NewLogrusLogger(logger)

	if flag.NArg() < 2 {
		logger.Error("Must have at least one input file and an output file specified.")
		os.Exit(2)
	}
	lastArg := flag.NArg() - 1

	// Open all the in-files for reading
	var readers []io.Reader
	consumingStdin := false
	for _, inFilename := range flag.Args()[:lastArg] {
		if inFilename == "-" && !consumingStdin {
			logger.Info("Reading from stdin")
			// Only add stdin once.
			consumingStdin = true
			readers = append(readers, os.Stdin)
			continue
		}
		logger.WithFields(log.Fields{
			"filename": inFilename,
		}).Debug("Reading from file")
		f, err := os.Open(inFilename)
		if err != nil {
			logger.WithFields(log.Fields{
				"error":    err,
				"filename": inFilename,
			}).Error("Failed to open an input file")
			os.Exit(2)
		}
		defer f.Close()
		readers = append(readers, f)
	}
	r := io.MultiReader(readers...)

	// Open the out-file for writing
	outFilename := flag.Args()[lastArg]
	w, err := outfile.Open(outFilename)
	if err != nil {
		logger.WithFields(log.Fields{
			"error":    err,
			"filename": outFilename,
		}).Error("Failed to open file for output")
		os.Exit(2)
	}
	defer w.Close()

	ctx := context.Background()

	// Prepare a client for communicating with GitHub's GraphQLv4 API and Restv3 API
	rt := roundtripper.NewTransport(ctx, scLogger)
	httpClient := &http.Client{
		Transport: rt,
	}
	ghClient := githubapi.NewClient(httpClient)

	ghRepoFactory := github.NewRepoFactory(ghClient, logger)

	// Register all the collectors that are supported.
	collector.Register(&github.RepoCollector{})
	collector.Register(&github.IssuesCollector{})
	collector.Register(githubmentions.NewCollector(ghClient))

	// Prepare the output writer
	out := result.NewCsvWriter(w, collector.EmptySets())

	// Read in each line from the input files
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		u, err := url.Parse(strings.TrimSpace(line))
		if err != nil {
			logger.WithFields(log.Fields{
				"error": err,
				"url":   line,
			}).Error("Failed to parse project url")
			os.Exit(1) // TODO: add a flag to continue or abort on failure
		}
		logger.WithFields(log.Fields{
			"url": u.String(),
		}).Debug("Parsed project url")

		// TODO: Create a resolver in the project package.
		r, err := ghRepoFactory.New(ctx, u)
		if err != nil {
			logger.WithFields(log.Fields{
				"error": err,
				"url":   u.String(),
			}).Warning("Failed to create project")
			continue
			//os.Exit(1) // TODO: add a flag to continue or abort on failure
			// TODO: we should have an error that indicates that the URL/Project
			// should be skipped/ignored.
		}

		// TODO: p.URL() should be checked to see if it has already been processed.

		// Collect the signals for the given project
		logger.WithFields(log.Fields{
			"url": u.String(),
		}).Info("Collecting")
		ss, err := collector.Collect(ctx, r)
		if err != nil {
			logger.WithFields(log.Fields{
				"error": err,
				"url":   r.URL().String(),
			}).Error("Failed to collect signals for project")
			os.Exit(1) // TODO: add a flag to continue or abort on failure
		}

		rec := out.Record()
		for _, s := range ss {
			if err := rec.WriteSignalSet(s); err != nil {
				logger.WithFields(log.Fields{
					"error": err,
					"url":   r.URL().String(),
				}).Error("Failed to write signal set")
				os.Exit(1) // TODO: add a flag to continue or abort on failure
			}
		}
		if err := rec.Done(); err != nil {
			logger.WithFields(log.Fields{
				"error":      err,
				"repository": line,
			}).Error("Failed to complete record")
			os.Exit(1) // TODO: add a flag to continue or abort on failure
		}
	}
	if err := scanner.Err(); err != nil {
		logger.WithFields(log.Fields{
			"error": err,
		}).Error("Failed while reading input")
		os.Exit(2)
	}

	// TODO: track metrics as we are running to measure coverage of data
}

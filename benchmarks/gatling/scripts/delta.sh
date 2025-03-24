#!/bin/bash

rm -rf serverpersecondloadsimulation-delta
java -jar ~/Downloads/gatling-report-6.3-capsule-fat.jar $(ls ~/Folders/Stuff/university/7sem/test/7sem-testing/benchmarks/gatling/results/serverpersecondloadsimulation*/simulation.log | tail -n 2) -o serverpersecondloadsimulation-delta

rm -rf serveratonceloadsimulation-delta
java -jar ~/Downloads/gatling-report-6.3-capsule-fat.jar $(ls ~/Folders/Stuff/university/7sem/test/7sem-testing/benchmarks/gatling/results/serveratonceloadsimulation*/simulation.log | tail -n 2) -o serveratonceloadsimulation-delta

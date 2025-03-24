#!/bin/bash

rm -rf serverpersecondloadsimulation-trend
java -jar ~/Downloads/gatling-report-6.3-capsule-fat.jar $(ls ~/Folders/Stuff/university/7sem/test/7sem-testing/benchmarks/gatling/results/serverpersecondloadsimulation*/simulation.log) -o serverpersecondloadsimulation-trend

rm -rf serveratonceloadsimulation-trend
java -jar ~/Downloads/gatling-report-6.3-capsule-fat.jar $(ls ~/Folders/Stuff/university/7sem/test/7sem-testing/benchmarks/gatling/results/serveratonceloadsimulation*/simulation.log) -o serveratonceloadsimulation-trend

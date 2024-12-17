package main

import (
	"barnes-hut-parallel/src/barneshut"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"

	// "proj3-redesigned/barneshut"
	"runtime"
	"strconv"
	"time"
)

func main() {

	// LEFT THIS FOR PROFILING IN FUTURE
	// // CPU PROFILING
	// f, err := os.Create("cpu.prof")
	// if err != nil {
	// 	fmt.Println("Could not create CPU profile:", err)
	// 	return
	// }
	// defer f.Close()
	// if err := pprof.StartCPUProfile(f); err != nil {
	// 	fmt.Println("Could not start CPU profile:", err)
	// 	return
	// }
	// defer pprof.StopCPUProfile()

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Number of particles
	nParticles := 10000
	numThreads := 1
	nIters := 200
	dt := 1.0
	args := os.Args
	var visual bool
	if len(args) > 1 {
		val, err := strconv.Atoi(args[1])
		if err == nil {
			nParticles = val
		}
		// Number of threads
		if len(args) > 2 {
			val, err := strconv.Atoi(args[2])
			if err == nil {
				numThreads = val
			}
		}
		// Iterations
		if len(args) > 3 {
			val, err := strconv.Atoi(args[3])
			if err == nil {
				nIters = val
			}
		}
		// Visual
		if len(args) > 4 {
			val := args[4]
			if val == "y" {
				visual = true
			} else {
				visual = false
			}
		}
	}

	// Create particles
	particles := make([]*barneshut.Particle, nParticles)
	for i := 0; i < nParticles; i++ {
		x := (rand.Float64() * 20000.0) - 10000 // random in [-1,1]
		y := (rand.Float64() * 20000.0) - 10000
		p := barneshut.NewParticle(x, y)
		particles[i] = p
	}

	// Create root node
	root := barneshut.CreateNode(float64(math.MinInt64), float64(math.MaxInt64), float64(math.MinInt64), float64(math.MaxInt64), nil)

	// Insert particles into the tree
	for i := 0; i < nParticles; i++ {
		barneshut.InsertParticle(root, particles[i])
	}

	// Open file for writing particle data
	datafile_input, err := os.Create("particles_input.dat")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer datafile_input.Close()

	barneshut.FprintDataFile(datafile_input, root)

	// Open file for writing particle data
	datafile, err := os.Create("particles_output.dat")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer datafile.Close()

	if visual {
		// Running in visual mode.
		// Start the python realtime plotter.
		cmd := exec.Command("python3", "space_graph.py", "particles_output.dat", strconv.Itoa(nParticles))
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Start()
		if err != nil {
			panic(err)
		}
		defer cmd.Process.Kill()
	}

	// Main loop
	startTime := time.Now()
	for iter := 1; iter <= nIters; iter++ {
		// fmt.Printf("iteration:%d\n", iter)
		newRoot := barneshut.CreateNode(float64(math.MinInt64), float64(math.MaxInt64), float64(math.MinInt64), float64(math.MaxInt64), nil)
		// Run the N-Body Simulation
		barneshut.RunSimulation(root, numThreads, dt, nParticles)
		// Recreate the tree with new positons
		barneshut.RecreateWithNewPos(root, newRoot)
		root = newRoot

		if visual {
			// Truncate the file to clear old contents
			if err := datafile.Truncate(0); err != nil {
				fmt.Println("Error truncating datafile:", err)
				return
			}

			// Reset the file offset to the start
			if _, err := datafile.Seek(0, 0); err != nil {
				fmt.Println("Error seeking datafile:", err)
				return
			}
			barneshut.FprintDataFile(datafile, root)
		}
	}
	endTime := time.Now()
	elapsedTime := endTime.Sub(startTime)
	barneshut.FprintDataFile(datafile, root)
	fmt.Println(elapsedTime.Seconds())
}

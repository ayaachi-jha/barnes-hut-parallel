import matplotlib.pyplot as plt
import subprocess

testRepeat = 1

requestSizes = [10000, 50000, 100000]
threads = [2, 4, 6, 8, 12]

sequentialTime = dict()

# Loop through request sizes and store the sequential runtimes
for requestSize in requestSizes :
    for i in range(testRepeat):
        result = subprocess.check_output(['go', 'run', 'main.go', str(requestSize)])
        if requestSize not in sequentialTime:
            sequentialTime[requestSize] = 0.000
        sequentialTime[requestSize] = sequentialTime[requestSize] + float(result.decode('utf-8'))
    sequentialTime[requestSize] = sequentialTime[requestSize]/testRepeat
    print(f'Sequential Time for {requestSize}: {sequentialTime[requestSize]}')

speedup = dict()

# Loop through request sizes and threads and find the speedups for each thread count
for requestSize in requestSizes :
    for thread in threads:
        time = 0.0
        for i in range(testRepeat):
            result = subprocess.check_output(['go', 'run', 'main.go', str(requestSize), str(thread)])
            time += float(result.decode('utf-8'))
        time = time/testRepeat
        speedupForThread = sequentialTime[requestSize]/time
        if requestSize not in speedup:
            speedup[requestSize] = {}
        speedup[requestSize][thread] = speedupForThread
        print(f'Speedup for {requestSize} particles with {thread} threads: {speedup[requestSize][thread]}')

# Plot graphs and store in speedup-graph.png
for requestSize in requestSizes:
    y1 = []
    for thread in threads:
        y1.append(speedup[requestSize][thread])
    labelName = str(requestSize) + " Particles"
    plt.plot(threads, y1, label=labelName)

plot_title = "Speedup Graph for Barnes Hut Algorithm"

plt.xlabel("No. of Threads")
plt.ylabel("Speedup")
plt.title(plot_title)
plt.legend()
# plt.show()
plt.savefig('speedup-graph.png') 
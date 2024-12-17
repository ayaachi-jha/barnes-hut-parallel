import matplotlib.pyplot as plt
import os
import time
import sys

def read_data(filename, expected_lines):
    x, y = [], []
    try:
        # Check the number of lines first
        with open(filename, 'r') as f:
            lines = f.readlines()
            if len(lines) != expected_lines:
                # File in between updation so don't read.
                return x, y 
            
            for line in lines:
                data = line.strip().split()
                if len(data) == 2:
                    x.append(float(data[0]))
                    y.append(float(data[1]))
    except Exception as e:
        print("Error reading file:", e)
    return x, y

def update_plot(fig, ax, filename, expected_lines, xlim=(-20000, 20000), ylim=(-20000, 20000)):
    x, y = read_data(filename, expected_lines)
    if not x or not y:  # Skip update if no data was read
        return

    ax.clear()  # Clear the previous plot
    
    # Plot the new particle positions
    ax.scatter(x, y, label="Particles", color='tab:blue')

    # Fix the axis limits (no dynamic adjustment)
    ax.set_xlim(xlim)
    ax.set_ylim(ylim)

    # Plot settings
    ax.set_title("Real-Time Particle Position Updates")
    ax.set_xlabel("X-Coordinate")
    ax.set_ylabel("Y-Coordinate")
    ax.legend()

    fig.canvas.draw()
    plt.pause(0.005)

def monitor_file(filename, expected_lines, interval=0.5):
    if not os.path.isfile(filename):
        print(f"File {filename} not found.")
        return

    fig, ax = plt.subplots()
    plt.ion()  # Turn on interactive mode
    plt.show(block=False)
    
    last_mod_time = 0
    while True:
        try:
            # Check if the file has been updated
            mod_time = os.path.getmtime(filename)
            if mod_time != last_mod_time:
                last_mod_time = mod_time
                update_plot(fig, ax, filename, expected_lines)
            
            plt.pause(interval)  # Allow the UI to refresh
        except KeyboardInterrupt:
            break
        except Exception as e:
            print("Error:", e)
            time.sleep(interval)

if __name__ == "__main__":
    # Get the filename and expected number of particles from command-line arguments
    if len(sys.argv) != 3:
        print("Usage: python script.py <filename> <expected_lines>")
        sys.exit(1)

    filename = sys.argv[1]
    try:
        expected_lines = int(sys.argv[2])
    except ValueError:
        print("Error: expected_lines must be an integer.")
        sys.exit(1)

    monitor_file(filename, expected_lines, interval=1)

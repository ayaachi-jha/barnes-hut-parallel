# !/bin/bash

# SBATCH --mail-user=ayaachi@cs.uchicago.edu
# SBATCH --mail-type=ALL
# SBATCH --output=/home/ayaachi/slurm/out/%j.%N.stdout
# SBATCH --error=/home/ayaachi/slurm/out/%j.%N.stderr
# SBATCH --chdir=/home/ayaachi/proj3/main
# SBATCH --partition=general
# SBATCH --job-name=generate_graph
# SBATCH --nodes=1
# SBATCH --ntasks=1
# SBATCH --cpus-per-task=16
# SBATCH --mem-per-cpu=1500
# SBATCH --exclusive
# SBATCH --time=200:00

python generate_graphs.py
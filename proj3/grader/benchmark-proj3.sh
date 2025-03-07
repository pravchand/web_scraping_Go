#!/bin/bash
#
#SBATCH --mail-user=praveenc@uchicago.edu
#SBATCH --mail-type=ALL
#SBATCH --job-name=proj3_benchmark 
#SBATCH --output=./slurm/out/%j.%N.stdout
#SBATCH --error=./slurm/out/%j.%N.stderr
#SBATCH --chdir=/home/praveenc/project-3-pravchand/proj3/benchmark
#SBATCH --partition=debug 
#SBATCH --nodes=1
#SBATCH --ntasks=1
#SBATCH --cpus-per-task=16
#SBATCH --mem-per-cpu=900
#SBATCH --exclusive
#SBATCH --time=60:00

module load golang/1.19

test_sizes=(50 100 200 400)
thread_counts=(2 4 6 8 12)
work_stealing_options=(true false)
runs=5



echo "Running benchmark tests with $(date)"

echo "Running sequential tests..."
for size in "${test_sizes[@]}"; do
    for ((i=1; i<=runs; i++)); do
        output=$(go run /home/praveenc/project-3-pravchand/proj3/parse/parse.go ${size})
        time=$(echo "$output" | grep "BENCHMARK_TIME:" | awk '{print $2}')
        echo "$time" >> "benchmark_results/sequential_${size}.txt"
    done
done

for size in "${test_sizes[@]}"; do
    for threads in "${thread_counts[@]}"; do
        for stealing in "${work_stealing_options[@]}"; do
            for ((i=1; i<=runs; i++)); do
                output=$(go run /home/praveenc/project-3-pravchand/proj3/parse/parse.go ${size} ${threads} ${stealing})
                time=$(echo "$output" | grep "BENCHMARK_TIME:" | awk '{print $2}')
                echo "$time" >> "benchmark_results/parallel_${size}_${threads}_${stealing}.txt"
            done
        done
    done
done

cat > plot_speedup.py << 'EOF'
import matplotlib.pyplot as plt
import numpy as np
import glob
import os

def read_times(filename):
    with open(filename) as f:
        return np.mean([float(line.strip()) for line in f])

# Read results
test_sizes = [50, 100, 200, 400]
thread_counts = [2, 4, 6, 8, 12]
work_stealing_options = ['true', 'false']

# Calculate speedups for both work stealing configurations
for stealing in work_stealing_options:
    plt.figure(figsize=(12, 8))
    
    for size in test_sizes:
        # Read sequential time
        seq_time = read_times(f'benchmark_results/sequential_{size}.txt')
        
        # Calculate speedups for each thread count
        speedups = []
        for threads in thread_counts:
            par_time = read_times(f'benchmark_results/parallel_{size}_{threads}_{stealing}.txt')
            speedup = seq_time / par_time
            speedups.append(speedup)
        
        # Plot speedup line
        plt.plot(thread_counts, speedups, 'o-', label=f'{size} links')
    
    # Add ideal speedup line (y=x)
    plt.plot(thread_counts, thread_counts, 'k--', label='Ideal Speedup')
    
    plt.title(f'Speedup vs. Threads (Work Stealing: {stealing})')
    plt.xlabel('Number of Threads')
    plt.ylabel('Speedup')
    plt.grid(True)
    plt.legend()
    plt.savefig(f'speedup_workstealing_{stealing}.png')

# Compare work stealing true vs false for each test size
for size in test_sizes:
    plt.figure(figsize=(12, 8))
    
    seq_time = read_times(f'benchmark_results/sequential_{size}.txt')
    
    for stealing in work_stealing_options:
        speedups = []
        for threads in thread_counts:
            par_time = read_times(f'benchmark_results/parallel_{size}_{threads}_{stealing}.txt')
            speedup = seq_time / par_time
            speedups.append(speedup)
        
        plt.plot(thread_counts, speedups, 'o-', label=f'Work Stealing: {stealing}')
    
    # Add ideal speedup line
    plt.plot(thread_counts, thread_counts, 'k--', label='Ideal Speedup')
    
    plt.title(f'Work Stealing Comparison ({size} links)')
    plt.xlabel('Number of Threads')
    plt.ylabel('Speedup')
    plt.grid(True)
    plt.legend()
    plt.savefig(f'work_stealing_comparison_{size}.png')

print("Plots generated")
EOF

echo "Generating speedup plots..."
python3 plot_speedup.py

echo "Benchmark completed"
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns

# Read in the saved CSV data.
benchmark_data = pd.read_csv('result.csv', header=0, names=['name', 'time', 'range'])

# Go stores benchmark results in nanoseconds. Convert all results to seconds.
benchmark_data['time'] /= 1e+9

# Use the name of the benchmark to extract the number of worker threads used.
#  e.g. "Filter/16-8" used 16 worker threads (goroutines).
# Note how the benchmark name corresponds to the regular expression 'Filter/\d+_workers-\d+'.
# Also note how we place brackets around the value we want to extract.
#benchmark_data['num_Workers'] = benchmark_data['name'].str.extract('Gol/\d+x\d+x\d+x(\d+)x\d+-\d+').apply(pd.to_numeric)
#benchmark_data['cpu_cores'] = benchmark_data['name'].str.extract('Gol/\d+x\d+x\d+x\d+x\d+-(\d+)').apply(pd.to_numeric)
#benchmark_data['worker_cores'] = benchmark_data['name'].str.extract('Gol/\d+x\d+x\d+x\d+x(\d+)-\d+').apply(pd.to_numeric)
benchmark_data['way'] = benchmark_data['name'].str.extract('Gol/\d+-(\d+)-\w+').apply(pd.to_numeric)
print(benchmark_data)



# Plot a bar chart.
ax = sns.barplot(data=benchmark_data, x='way', y='time')

# Set descriptive axis lables.
ax.set(xlabel='Way of transport data', ylabel='Time taken (s)')

# Display the full figure.
plt.show()

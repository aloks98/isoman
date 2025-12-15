import { useMemo } from 'react';
import { Cell, Pie, PieChart } from 'recharts';
import {
  type ChartConfig,
  ChartContainer,
  ChartLegend,
  ChartLegendContent,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart';

interface DistributionChartProps {
  data: Record<string, number>;
  title?: string;
  colorByKey?: boolean;
}

const DEFAULT_COLORS = [
  'var(--chart-1)',
  'var(--chart-2)',
  'var(--chart-3)',
  'var(--chart-4)',
  'var(--chart-5)',
];

const STATUS_COLORS: Record<string, string> = {
  complete: 'var(--chart-2)',
  completed: 'var(--chart-2)',
  failed: 'var(--chart-4)',
  error: 'var(--chart-4)',
  pending: 'var(--chart-3)',
  queued: 'var(--chart-3)',
  downloading: 'var(--chart-1)',
  in_progress: 'var(--chart-1)',
  verifying: 'var(--chart-5)',
};

export function DistributionChart({
  data,
  colorByKey = false,
}: DistributionChartProps) {
  const chartData = useMemo(() => {
    return Object.entries(data)
      .sort((a, b) => b[1] - a[1])
      .map(([name, value]) => ({
        name: name.charAt(0).toUpperCase() + name.slice(1),
        value,
        originalKey: name.toLowerCase(),
      }));
  }, [data]);

  const chartConfig = useMemo(() => {
    const config: ChartConfig = {};
    chartData.forEach((entry, index) => {
      const color =
        colorByKey && STATUS_COLORS[entry.originalKey]
          ? STATUS_COLORS[entry.originalKey]
          : DEFAULT_COLORS[index % DEFAULT_COLORS.length];
      config[entry.name] = {
        label: entry.name,
        color,
      };
    });
    return config;
  }, [chartData, colorByKey]);

  if (chartData.length === 0) {
    return (
      <div className="text-muted-foreground text-center py-8">
        No data available
      </div>
    );
  }

  return (
    <ChartContainer config={chartConfig} className="h-[280px] w-full">
      <PieChart>
        <ChartTooltip content={<ChartTooltipContent nameKey="name" />} />
        <Pie
          data={chartData}
          cx="50%"
          cy="45%"
          innerRadius={60}
          outerRadius={90}
          paddingAngle={2}
          dataKey="value"
          nameKey="name"
        >
          {chartData.map((entry) => (
            <Cell key={entry.name} fill={chartConfig[entry.name]?.color} />
          ))}
        </Pie>
        <ChartLegend content={<ChartLegendContent nameKey="name" />} />
      </PieChart>
    </ChartContainer>
  );
}

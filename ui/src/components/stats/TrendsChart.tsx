import { useMemo } from 'react';
import { Bar, BarChart, CartesianGrid, XAxis, YAxis } from 'recharts';
import {
  type ChartConfig,
  ChartContainer,
  ChartTooltip,
  ChartTooltipContent,
} from '@/components/ui/chart';
import type { TrendDataPoint } from '@/types/stats';

interface TrendsChartProps {
  data: TrendDataPoint[];
}

const chartConfig = {
  count: {
    label: 'Downloads',
    color: 'var(--chart-5)',
  },
} satisfies ChartConfig;

export function TrendsChart({ data }: TrendsChartProps) {
  const chartData = useMemo(() => {
    return data.map((d) => {
      const parts = d.date.split('-');
      return {
        ...d,
        label: parts.length >= 2 ? `${parts[1]}/${parts[2] || ''}` : d.date,
      };
    });
  }, [data]);

  if (data.length === 0) {
    return (
      <div className="text-muted-foreground text-center py-8">
        No download data yet
      </div>
    );
  }

  return (
    <ChartContainer config={chartConfig} className="h-[300px] w-full">
      <BarChart
        data={chartData}
        margin={{ top: 10, right: 10, left: 10, bottom: 30 }}
      >
        <CartesianGrid strokeDasharray="3 3" vertical={false} />
        <XAxis
          dataKey="label"
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tick={{ fontSize: 11 }}
          label={{
            value: 'Date',
            position: 'insideBottom',
            offset: -15,
            className: 'fill-muted-foreground text-xs',
          }}
        />
        <YAxis
          tickLine={false}
          axisLine={false}
          tickMargin={8}
          tick={{ fontSize: 11 }}
          label={{
            value: 'Downloads',
            angle: -90,
            position: 'insideLeft',
            style: { textAnchor: 'middle' },
            className: 'fill-muted-foreground text-xs',
          }}
        />
        <ChartTooltip
          content={
            <ChartTooltipContent
              hideLabel={false}
              labelFormatter={(value, payload) => {
                const date = payload?.[0]?.payload?.date;
                return date ? `Date: ${date}` : value;
              }}
            />
          }
        />
        <Bar
          dataKey="count"
          fill="var(--color-count)"
          radius={[4, 4, 0, 0]}
          maxBarSize={50}
        />
      </BarChart>
    </ChartContainer>
  );
}

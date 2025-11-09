import { useState } from 'react';
import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';
import type { WidgetProps } from './widget-registry';

interface CostCategory {
  name: string;
  value: number;
  color: string;
  [key: string]: string | number;
}

export function CostBreakdownWidget({ config, onOpenModal }: WidgetProps) {
  const timeRange = (config.time_range as string) || 'month';

  const [selectedCategory, setSelectedCategory] = useState<string | null>(null);

  const mockData: CostCategory[] = [
    { name: 'Feature Development', value: 125000, color: '#3B82F6' },
    { name: 'Technical Debt', value: 45000, color: '#F59E0B' },
    { name: 'Support & Maintenance', value: 32000, color: '#10B981' },
    { name: 'Operations', value: 18000, color: '#8B5CF6' },
    { name: 'Research & Innovation', value: 15000, color: '#EC4899' },
  ];

  const totalCost = mockData.reduce((sum, item) => sum + item.value, 0);

  const formatCurrency = (value: number) => {
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 0,
      maximumFractionDigits: 0,
    }).format(value);
  };

  const handleSegmentClick = (category: string) => {
    setSelectedCategory(category === selectedCategory ? null : category);
    if (onOpenModal) {
      onOpenModal('cost-detail', { category, timeRange });
    }
  };

  return (
    <div className="bg-white dark:bg-gray-800 border border-gray-200 dark:border-gray-700 rounded-lg p-6 h-full flex flex-col">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-sm font-medium text-gray-600 dark:text-gray-400">Cost Breakdown</h3>
        <div className="text-right">
          <p className="text-xs text-gray-500 dark:text-gray-400">Total Cost</p>
          <p className="text-lg font-bold text-gray-900 dark:text-gray-100">
            {formatCurrency(totalCost)}
          </p>
        </div>
      </div>

      <div className="flex-1 min-h-0 mb-4">
        <ResponsiveContainer width="100%" height="100%">
          <PieChart>
            <Pie
              data={mockData}
              cx="50%"
              cy="50%"
              innerRadius="60%"
              outerRadius="80%"
              paddingAngle={2}
              dataKey="value"
              onClick={(data: any) => handleSegmentClick(data.name)}
              cursor="pointer"
            >
              {mockData.map((entry, index) => (
                <Cell
                  key={`cell-${index}`}
                  fill={entry.color}
                  opacity={selectedCategory === null || selectedCategory === entry.name ? 1 : 0.3}
                />
              ))}
            </Pie>
            <Tooltip
              formatter={(value: number) => formatCurrency(value)}
              contentStyle={{
                backgroundColor: 'rgba(0, 0, 0, 0.8)',
                border: 'none',
                borderRadius: '8px',
                color: '#fff',
              }}
            />
            <Legend
              verticalAlign="bottom"
              height={36}
              formatter={(value) => (
                <span className="text-xs text-gray-700 dark:text-gray-300">{value}</span>
              )}
            />
          </PieChart>
        </ResponsiveContainer>
      </div>

      <div className="space-y-2 pt-4 border-t border-gray-200 dark:border-gray-700">
        {mockData.map((item) => {
          const percentage = ((item.value / totalCost) * 100).toFixed(1);
          const isSelected = selectedCategory === item.name;

          return (
            <div
              key={item.name}
              onClick={() => handleSegmentClick(item.name)}
              className={`flex items-center justify-between p-2 rounded cursor-pointer transition-colors ${
                isSelected
                  ? 'bg-gray-100 dark:bg-gray-900'
                  : 'hover:bg-gray-50 dark:hover:bg-gray-900/50'
              }`}
            >
              <div className="flex items-center gap-2">
                <div className="w-3 h-3 rounded-full" style={{ backgroundColor: item.color }} />
                <span className="text-sm text-gray-900 dark:text-gray-100">{item.name}</span>
              </div>
              <div className="text-right">
                <p className="text-sm font-medium text-gray-900 dark:text-gray-100">
                  {formatCurrency(item.value)}
                </p>
                <p className="text-xs text-gray-500 dark:text-gray-400">{percentage}%</p>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
}

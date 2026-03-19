import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
} from 'recharts';

const plSummary = [
  { label: 'Gross Revenue', value: '$4,250' },
  { label: 'Net Revenue', value: '$3,960' },
  { label: 'COGS', value: '$1,275' },
  { label: 'Gross Profit', value: '$2,685' },
  { label: 'Margin %', value: '67.8%' },
];

const channelData = [
  { channel: 'Dine-in', revenue: 2100, checks: 42, avgCheck: 50.0 },
  { channel: 'Takeout', revenue: 1050, checks: 35, avgCheck: 30.0 },
  { channel: 'Delivery', revenue: 1100, checks: 44, avgCheck: 25.0 },
];

export default function FinancialPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Financial Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          P&L overview and channel performance
        </p>
      </div>

      {/* P&L Summary Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-5 gap-4">
        {plSummary.map(({ label, value }) => (
          <div
            key={label}
            className="bg-white rounded-xl border border-gray-200 p-4 shadow-sm"
          >
            <p className="text-xs text-gray-500 uppercase tracking-wide">{label}</p>
            <p className="text-xl font-bold text-gray-800 mt-1">{value}</p>
          </div>
        ))}
      </div>

      {/* Channel Revenue Chart */}
      <div className="bg-white rounded-xl border border-gray-200 p-6 shadow-sm">
        <h2 className="text-lg font-semibold text-gray-800 mb-4">
          Revenue by Channel
        </h2>
        <div className="h-72">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart
              data={channelData}
              margin={{ top: 5, right: 20, left: 0, bottom: 5 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#E5E7EB" />
              <XAxis dataKey="channel" tick={{ fontSize: 13 }} />
              <YAxis
                tick={{ fontSize: 13 }}
                tickFormatter={(v: number) => `$${v.toLocaleString()}`}
              />
              <Tooltip
                formatter={(value) => [`$${Number(value).toLocaleString()}`, 'Revenue']}
                contentStyle={{
                  borderRadius: '8px',
                  border: '1px solid #E5E7EB',
                  fontSize: '13px',
                }}
              />
              <Bar dataKey="revenue" fill="#F97316" radius={[6, 6, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </div>

      {/* Channel Breakdown Table */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-800">
            Channel Breakdown
          </h2>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 text-left text-gray-500 uppercase tracking-wider text-xs">
                <th className="px-6 py-3 font-medium">Channel</th>
                <th className="px-6 py-3 font-medium text-right">Revenue</th>
                <th className="px-6 py-3 font-medium text-right">Checks</th>
                <th className="px-6 py-3 font-medium text-right">Avg Check</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {channelData.map(({ channel, revenue, checks, avgCheck }) => (
                <tr key={channel} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-3 font-medium text-gray-800">{channel}</td>
                  <td className="px-6 py-3 text-right text-gray-700">
                    ${revenue.toLocaleString()}
                  </td>
                  <td className="px-6 py-3 text-right text-gray-700">{checks}</td>
                  <td className="px-6 py-3 text-right text-gray-700">
                    ${avgCheck.toFixed(2)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

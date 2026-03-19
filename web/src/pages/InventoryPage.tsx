type ParStatus = 'OK' | 'Low' | 'Critical';

const usageData = [
  { ingredient: 'Ground Beef', qty: 18, unit: 'lbs', costPerUnit: 4.5, totalCost: 81.0 },
  { ingredient: 'Chicken Breast', qty: 14, unit: 'lbs', costPerUnit: 3.75, totalCost: 52.5 },
  { ingredient: 'Cheese', qty: 8, unit: 'lbs', costPerUnit: 5.2, totalCost: 41.6 },
  { ingredient: 'Lettuce', qty: 6, unit: 'heads', costPerUnit: 2.0, totalCost: 12.0 },
  { ingredient: 'Tomatoes', qty: 10, unit: 'lbs', costPerUnit: 2.8, totalCost: 28.0 },
];

const parData: {
  ingredient: string;
  current: number;
  parLevel: number;
  reorderPoint: number;
  status: ParStatus;
}[] = [
  { ingredient: 'Ground Beef', current: 12, parLevel: 30, reorderPoint: 15, status: 'Critical' },
  { ingredient: 'Chicken Breast', current: 22, parLevel: 25, reorderPoint: 12, status: 'Low' },
  { ingredient: 'Cheese', current: 18, parLevel: 15, reorderPoint: 8, status: 'OK' },
  { ingredient: 'Lettuce', current: 10, parLevel: 12, reorderPoint: 6, status: 'Low' },
  { ingredient: 'Tomatoes', current: 20, parLevel: 18, reorderPoint: 9, status: 'OK' },
];

const statusBadge: Record<ParStatus, string> = {
  OK: 'bg-emerald-100 text-emerald-700',
  Low: 'bg-yellow-100 text-yellow-700',
  Critical: 'bg-red-100 text-red-700',
};

export default function InventoryPage() {
  return (
    <div className="space-y-8">
      <div>
        <h1 className="text-2xl font-bold text-gray-800">Inventory Intelligence</h1>
        <p className="text-sm text-gray-500 mt-1">
          Theoretical usage and PAR status overview
        </p>
      </div>

      {/* Theoretical Usage */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-800">Theoretical Usage</h2>
          <p className="text-xs text-gray-500 mt-0.5">Based on today's sales mix</p>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 text-left text-gray-500 uppercase tracking-wider text-xs">
                <th className="px-6 py-3 font-medium">Ingredient</th>
                <th className="px-6 py-3 font-medium text-right">Qty Used</th>
                <th className="px-6 py-3 font-medium">Unit</th>
                <th className="px-6 py-3 font-medium text-right">Cost/Unit</th>
                <th className="px-6 py-3 font-medium text-right">Total Cost</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {usageData.map(({ ingredient, qty, unit, costPerUnit, totalCost }) => (
                <tr key={ingredient} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-3 font-medium text-gray-800">{ingredient}</td>
                  <td className="px-6 py-3 text-right text-gray-700">{qty}</td>
                  <td className="px-6 py-3 text-gray-500">{unit}</td>
                  <td className="px-6 py-3 text-right text-gray-700">
                    ${costPerUnit.toFixed(2)}
                  </td>
                  <td className="px-6 py-3 text-right text-gray-700">
                    ${totalCost.toFixed(2)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* PAR Status */}
      <div className="bg-white rounded-xl border border-gray-200 shadow-sm overflow-hidden">
        <div className="px-6 py-4 border-b border-gray-200">
          <h2 className="text-lg font-semibold text-gray-800">PAR Status</h2>
          <p className="text-xs text-gray-500 mt-0.5">Current stock vs. target levels</p>
        </div>
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gray-50 text-left text-gray-500 uppercase tracking-wider text-xs">
                <th className="px-6 py-3 font-medium">Ingredient</th>
                <th className="px-6 py-3 font-medium text-right">Current</th>
                <th className="px-6 py-3 font-medium text-right">PAR Level</th>
                <th className="px-6 py-3 font-medium text-right">Reorder Point</th>
                <th className="px-6 py-3 font-medium text-center">Status</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {parData.map(({ ingredient, current, parLevel, reorderPoint, status }) => (
                <tr key={ingredient} className="hover:bg-gray-50 transition-colors">
                  <td className="px-6 py-3 font-medium text-gray-800">{ingredient}</td>
                  <td className="px-6 py-3 text-right text-gray-700">{current}</td>
                  <td className="px-6 py-3 text-right text-gray-700">{parLevel}</td>
                  <td className="px-6 py-3 text-right text-gray-700">{reorderPoint}</td>
                  <td className="px-6 py-3 text-center">
                    <span
                      className={`inline-block text-xs font-semibold px-2.5 py-0.5 rounded-full ${statusBadge[status]}`}
                    >
                      {status}
                    </span>
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

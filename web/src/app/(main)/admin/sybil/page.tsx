'use client';

import { useEffect, useState } from 'react';
import { api } from '@/lib/api-client';

interface SuspectUser {
  id: string;
  phone_number: string;
  display_name: string;
  trust_status: string;
  trust_score: number;
}

export default function SybilReviewDashboard() {
  const [users, setUsers] = useState<SuspectUser[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchUsers = async () => {
    try {
      setLoading(true);
      const res = await api.get('/v1/admin/users/under-review');
      // Axios interceptor extracts data, but depending on how it's written it could be res.data or res
      setUsers(res?.data || res || []);
      setError(null);
    } catch (err: any) {
      setError(err?.response?.data?.message || 'Failed to loaded suspected users');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchUsers();
  }, []);

  const handleReview = async (id: string, action: 'ban' | 'restore') => {
    try {
      await api.post(`/v1/admin/users/${id}/review`, { action });
      setUsers((prev) => prev.filter((u) => u.id !== id));
    } catch (err: any) {
      alert(`Failed to ${action} user: ${err?.message}`);
    }
  };

  if (loading) return <div className="p-8">Loading suspected users...</div>;
  if (error) return <div className="p-8 text-red-500">Error: {error}</div>;

  return (
    <div className="container mx-auto p-4 md:p-8">
      <h1 className="text-3xl font-bold mb-6 text-gray-800 dark:text-gray-100">Sybil Review Dashboard</h1>
      
      <div className="mb-6">
        <p className="text-gray-600 dark:text-gray-400">
          Review accounts flagged by the automated Sybil detection engine.
        </p>
      </div>

      {users.length === 0 ? (
        <div className="bg-white dark:bg-gray-800 shadow rounded-lg p-6 text-center text-gray-500">
          No users currently under review. Great job!
        </div>
      ) : (
        <div className="bg-white dark:bg-gray-800 shadow shadow-sm rounded-lg overflow-hidden">
          <div className="overflow-x-auto">
            <table className="w-full text-left border-collapse">
              <thead>
                <tr className="bg-gray-50 dark:bg-gray-700 text-gray-700 dark:text-gray-300 border-b dark:border-gray-600">
                  <th className="py-3 px-4 font-semibold text-sm">Display Name</th>
                  <th className="py-3 px-4 font-semibold text-sm">Phone Number</th>
                  <th className="py-3 px-4 font-semibold text-sm">Trust Status</th>
                  <th className="py-3 px-4 font-semibold text-sm">Trust Score</th>
                  <th className="py-3 px-4 font-semibold text-sm text-right">Actions</th>
                </tr>
              </thead>
              <tbody>
                {users.map((user) => (
                  <tr key={user.id} className="border-b border-gray-100 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700/50 transition">
                    <td className="py-3 px-4 text-gray-800 dark:text-gray-200">{user.display_name}</td>
                    <td className="py-3 px-4 text-gray-600 dark:text-gray-400 text-sm font-mono">{user.phone_number}</td>
                    <td className="py-3 px-4">
                      <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-400">
                        {user.trust_status}
                      </span>
                    </td>
                    <td className="py-3 px-4">
                      <div className="flex items-center">
                        <div className="w-16 bg-gray-200 dark:bg-gray-600 rounded-full h-2 mr-2">
                          <div 
                            className={`h-2 rounded-full ${user.trust_score < 30 ? 'bg-red-500' : 'bg-yellow-500'}`} 
                            style={{ width: `${Math.max(0, Math.min(100, user.trust_score))}%` }}
                          ></div>
                        </div>
                        <span className="text-sm font-medium text-gray-700 dark:text-gray-300">{user.trust_score}</span>
                      </div>
                    </td>
                    <td className="py-3 px-4 text-right">
                      <button
                        onClick={() => handleReview(user.id, 'restore')}
                        className="inline-flex justify-center py-1.5 px-3 border border-transparent shadow-sm text-sm font-medium rounded-md text-white bg-green-600 hover:bg-green-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-green-500 mr-2"
                      >
                        Approve
                      </button>
                      <button
                        onClick={() => handleReview(user.id, 'ban')}
                        className="inline-flex justify-center py-1.5 px-3 border border-transparent shadow-sm text-sm font-medium rounded-md text-red-700 bg-red-100 hover:bg-red-200 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
                      >
                        Ban
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  );
}

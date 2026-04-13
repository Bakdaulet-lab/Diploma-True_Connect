'use client';

import { useState } from 'react';
import Link from 'next/link';
import { useAdminSystemStats, useAdminSearchUsers, useSybilClusters } from '@/hooks/admin';
import { api } from '@/lib/api-client';
import { LoadingScreen } from '@/components/ui/common';
import { BarChart, Bar, LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

// Types (defined inline since they may not be in types file yet)
interface AdminDashboardStats {
  total_active_users: number;
  matches_made_today: number;
  average_trust_score: number;
}

interface AdminUserRow {
  id: string;
  phone: string;
  email: string;
  display_name?: string;
  avatar_url?: string;
  trust_score: number;
  trust_status: string;
  verification_level: string;
  is_active: boolean;
  created_at: string;
}

interface SybilCluster {
  CommunityID: number;
  Size: number;
  ExternalConnections: number;
  SuspectUIDs: string[];
}

// Generate dummy historical data for charts
const generateDummyHistoricalData = () => {
  const data = [];
  for (let i = 30; i >= 0; i--) {
    const date = new Date();
    date.setDate(date.getDate() - i);
    data.push({
      date: date.toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
      activeUsers: Math.floor(Math.random() * 500) + 1200,
      matches: Math.floor(Math.random() * 150) + 50,
      avgTrust: Math.floor(Math.random() * 30) + 60,
    });
  }
  return data;
};

const StatCard = ({ label, value, icon }: { label: string; value: string | number; icon: string }) => (
  <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
    <div className="flex items-center justify-between">
      <div>
        <p className="text-gray-600 text-sm font-medium">{label}</p>
        <p className="text-3xl font-bold text-gray-900 mt-2">{value}</p>
      </div>
      <div className="text-4xl">{icon}</div>
    </div>
  </div>
);

export default function AdminDashboardPage() {
  const { data: stats, isLoading: statsLoading } = useAdminSystemStats();
  const { data: clusters, isLoading: clustersLoading } = useSybilClusters();
  
  const [searchQuery, setSearchQuery] = useState('');
  const [hasSearched, setHasSearched] = useState(false);
  const [banningUsers, setBanningUsers] = useState(new Set<string>());
  
  const { data: searchResults, isLoading: searchLoading } = useAdminSearchUsers(
    searchQuery,
    20,
    0
  );

  const chartData = generateDummyHistoricalData();
  const users = searchResults?.items || [];

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (searchQuery.trim()) {
      setHasSearched(true);
    }
  };

  const handleBanRestore = async (userId: string, action: 'ban' | 'restore') => {
    try {
      setBanningUsers((prev) => new Set(prev).add(userId));
      await api.post(`/v1/admin/users/${userId}/review`, { action });
      
      // Refresh search results
      setHasSearched(false);
      setSearchQuery('');
    } catch (err: any) {
      const errorMsg = err?.response?.data?.message || err?.message || `Failed to ${action} user`;
      alert(`Error: ${errorMsg}`);
    } finally {
      setBanningUsers((prev) => {
        const newSet = new Set(prev);
        newSet.delete(userId);
        return newSet;
      });
    }
  };

  if (statsLoading && clustersLoading) {
    return <LoadingScreen />;
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100">
      <div className="max-w-7xl mx-auto py-8 px-4 sm:px-6 lg:px-8">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div>
            <h1 className="text-4xl font-bold text-gray-900">📊 Admin Dashboard</h1>
            <p className="text-gray-600 mt-2">System analytics, user management, and Sybil cluster monitoring</p>
          </div>
          <Link 
            href="/admin/sybil"
            className="px-4 py-2 bg-indigo-600 hover:bg-indigo-700 text-white font-medium rounded-xl transition-colors text-sm"
          >
            View Sybil Review
          </Link>
        </div>

        {/* === SECTION 1: System Analytics === */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
            <span className="bg-blue-100 p-2 rounded-lg">📈</span> System Analytics
          </h2>

          {/* Stat Cards */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
            <StatCard 
              label="Active Users" 
              value={stats?.total_active_users || 0}
              icon="👥"
            />
            <StatCard 
              label="Matches Today" 
              value={stats?.matches_made_today || 0}
              icon="💕"
            />
            <StatCard 
              label="Avg Trust Score" 
              value={stats?.average_trust_score ? Math.floor(stats.average_trust_score) : 0}
              icon="⭐"
            />
          </div>

          {/* Chart */}
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">30-Day Trend</h3>
            <ResponsiveContainer width="100%" height={300}>
              <BarChart data={chartData}>
                <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
                <XAxis 
                  dataKey="date" 
                  tick={{ fontSize: 12, fill: '#6b7280' }}
                  angle={-45}
                  textAnchor="end"
                  height={80}
                />
                <YAxis tick={{ fontSize: 12, fill: '#6b7280' }} />
                <Tooltip 
                  contentStyle={{ 
                    backgroundColor: '#fff', 
                    border: '1px solid #e5e7eb',
                    borderRadius: '8px'
                  }}
                />
                <Legend />
                <Bar dataKey="activeUsers" fill="#3b82f6" name="Active Users" />
                <Bar dataKey="matches" fill="#ec4899" name="Matches" />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* === SECTION 2: CRM User Management === */}
        <div className="mb-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
            <span className="bg-green-100 p-2 rounded-lg">👤</span> User Management
          </h2>

          {/* Search Bar */}
          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 mb-6">
            <form onSubmit={handleSearch} className="flex gap-3">
              <input
                type="text"
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder="Search by display name, email, or phone..."
                className="flex-1 px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:border-transparent"
              />
              <button
                type="submit"
                className="px-6 py-2 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 transition-colors"
              >
                Search
              </button>
            </form>
          </div>

          {/* Users Table */}
          {hasSearched ? (
            <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
              {searchLoading ? (
                <div className="p-8 text-center text-gray-500">Loading users...</div>
              ) : users.length === 0 ? (
                <div className="p-8 text-center text-gray-500">No users found matching "{searchQuery}"</div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead>
                      <tr className="bg-gray-50 border-b border-gray-100">
                        <th className="py-3 px-4 text-left text-sm font-semibold text-gray-900">Display Name</th>
                        <th className="py-3 px-4 text-left text-sm font-semibold text-gray-900">Email</th>
                        <th className="py-3 px-4 text-left text-sm font-semibold text-gray-900">Phone</th>
                        <th className="py-3 px-4 text-left text-sm font-semibold text-gray-900">Trust Score</th>
                        <th className="py-3 px-4 text-left text-sm font-semibold text-gray-900">Status</th>
                        <th className="py-3 px-4 text-left text-sm font-semibold text-gray-900">Verification</th>
                        <th className="py-3 px-4 text-right text-sm font-semibold text-gray-900">Actions</th>
                      </tr>
                    </thead>
                    <tbody>
                      {users.map((user) => (
                        <tr key={user.id} className="border-b border-gray-100 hover:bg-gray-50 transition-colors">
                          <td className="py-3 px-4 text-sm font-medium text-gray-900">
                            {user.display_name || '—'}
                          </td>
                          <td className="py-3 px-4 text-sm text-gray-600">{user.email}</td>
                          <td className="py-3 px-4 text-sm text-gray-600 font-mono">{user.phone}</td>
                          <td className="py-3 px-4">
                            <div className="flex items-center gap-2">
                              <div className="w-12 bg-gray-200 rounded-full h-2">
                                <div 
                                  className={`h-2 rounded-full ${
                                    user.trust_score < 30 ? 'bg-red-500' : 
                                    user.trust_score < 60 ? 'bg-yellow-500' : 
                                    'bg-green-500'
                                  }`}
                                  style={{ width: `${Math.max(0, Math.min(100, user.trust_score))}%` }}
                                ></div>
                              </div>
                              <span className="text-sm font-semibold text-gray-900">{user.trust_score}</span>
                            </div>
                          </td>
                          <td className="py-3 px-4">
                            <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                              user.is_active 
                                ? 'bg-green-100 text-green-800' 
                                : 'bg-red-100 text-red-800'
                            }`}>
                              {user.is_active ? 'Active' : 'Inactive'}
                            </span>
                          </td>
                          <td className="py-3 px-4">
                            <span className="text-xs font-medium text-gray-700 bg-blue-50 px-2.5 py-0.5 rounded">
                              {user.verification_level || 'None'}
                            </span>
                          </td>
                          <td className="py-3 px-4 text-right">
                            <div className="flex justify-end gap-2">
                              {user.is_active ? (
                                <button
                                  onClick={() => handleBanRestore(user.id, 'ban')}
                                  disabled={banningUsers.has(user.id)}
                                  className="px-3 py-1.5 bg-red-100 hover:bg-red-200 text-red-700 text-xs font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                  {banningUsers.has(user.id) ? 'Banning...' : 'Ban'}
                                </button>
                              ) : (
                                <button
                                  onClick={() => handleBanRestore(user.id, 'restore')}
                                  disabled={banningUsers.has(user.id)}
                                  className="px-3 py-1.5 bg-green-100 hover:bg-green-200 text-green-700 text-xs font-medium rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
                                >
                                  {banningUsers.has(user.id) ? 'Restoring...' : 'Restore'}
                                </button>
                              )}
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              )}
            </div>
          ) : (
            <div className="bg-white rounded-2xl shadow-sm border border-gray-100 p-8 text-center text-gray-500">
              Enter a search query and click "Search" to find users
            </div>
          )}
        </div>

        {/* === SECTION 3: Sybil Clusters === */}
        <div>
          <h2 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
            <span className="bg-red-100 p-2 rounded-lg">🛡️</span> Detected Sybil Clusters
          </h2>

          <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
            {clusters && clusters.length > 0 ? (
              <ul className="divide-y divide-gray-100">
                {clusters.map((cluster: SybilCluster) => (
                  <li key={cluster.CommunityID} className="p-6 hover:bg-gray-50 transition-colors">
                    <div className="flex justify-between items-start mb-4">
                      <div>
                        <span className="inline-flex items-center gap-1.5 px-3 py-1 rounded-full text-sm font-medium bg-red-100 text-red-800 mb-2">
                          Cluster #{cluster.CommunityID}
                        </span>
                        <h3 className="text-lg font-bold text-gray-900 mb-1">
                          {cluster.Size} Suspect Accounts
                        </h3>
                        <p className="text-sm text-gray-500">
                          External Connections: <b>{cluster.ExternalConnections}</b> (Low ratio typical of bot nets)
                        </p>
                      </div>
                      <button className="px-4 py-2 bg-red-600 hover:bg-red-700 text-white font-medium rounded-xl transition-colors text-sm">
                        Ban Cluster
                      </button>
                    </div>
                    <div className="bg-gray-100 rounded-lg p-4">
                      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">Account UUIDs in Cluster</p>
                      <div className="space-y-1 max-h-40 overflow-y-auto">
                        {cluster.SuspectUIDs.map((uid: string) => (
                          <div key={uid} className="text-xs font-mono text-gray-700">{uid}</div>
                        ))}
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            ) : (
              <div className="p-8 text-center text-gray-500">
                {clustersLoading ? 'Loading clusters...' : '🎉 No suspicious clusters found!'}
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
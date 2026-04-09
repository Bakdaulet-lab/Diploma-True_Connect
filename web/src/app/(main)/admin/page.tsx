'use client';

import { useSybilClusters } from '@/hooks/admin';
import { LoadingScreen } from '@/components/ui/common';

export default function AdminDashboardPage() {
  const { data: clusters, isLoading, error } = useSybilClusters();

  if (isLoading) return <LoadingScreen />;

  return (
    <div className="max-w-5xl mx-auto py-8 px-4">
      <h1 className="text-3xl font-bold mb-6 text-gray-900">🛡️ Admin Dashboard</h1>
      <p className="text-gray-600 mb-8">Monitor graph integrity and trust abuse clusters caught by the Neo4j Sybil detection algorithm.</p>

      <div className="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
        <div className="p-6 border-b border-gray-100 bg-red-50/30">
          <h2 className="text-xl font-bold text-red-900 flex items-center gap-2">
            <span className="bg-red-100 p-1.5 rounded-lg">🤖</span> Detected Sybil Clusters
          </h2>
        </div>

        {error ? (
          <div className="p-6 text-red-600">Failed to load sybil clusters. Are you an admin?</div>
        ) : !clusters || clusters.length === 0 ? (
          <div className="p-8 text-center text-gray-500">🎉 No suspicious clusters found!</div>
        ) : (
          <ul className="divide-y divide-gray-100">
            {clusters.map((cluster) => (
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
                  <div className="space-y-1">
                    {cluster.SuspectUIDs.map(uid => (
                      <div key={uid} className="text-xs font-mono text-gray-700">{uid}</div>
                    ))}
                  </div>
                </div>
              </li>
            ))}
          </ul>
        )}
      </div>
    </div>
  );
}
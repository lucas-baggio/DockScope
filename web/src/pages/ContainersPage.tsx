import { useEffect, useState } from 'react';
import { Eye, Loader2 } from 'lucide-react';
import { api } from '../services/api';
import type { Container } from '../types/docker';
import { ContainerDetailsModal } from '../components/ContainerDetailsModal';

export function ContainersPage() {
  const [containers, setContainers] = useState<Container[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedId, setSelectedId] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    api
      .getContainers(true)
      .then((data) => {
        if (!cancelled) setContainers(data);
      })
      .catch((err) => {
        if (!cancelled) setError(err instanceof Error ? err.message : 'Erro');
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const isRunning = (c: Container) =>
    c.State?.toLowerCase() === 'running';

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <Loader2 className="w-8 h-8 animate-spin text-zinc-400" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="p-6">
        <div className="rounded-lg bg-red-500/10 border border-red-500/30 text-red-400 px-4 py-3">
          {error}
        </div>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-semibold text-white">Containers</h1>
        <p className="text-zinc-400 text-sm mt-1">
          {containers.length} container(s)
        </p>
      </div>

      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-zinc-800">
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Nome
                </th>
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Imagem
                </th>
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Status
                </th>
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider w-32">
                  Ações
                </th>
              </tr>
            </thead>
            <tbody>
              {containers.map((c) => (
                <tr
                  key={c.ID}
                  className="border-b border-zinc-800/80 hover:bg-zinc-800/30 transition-colors"
                >
                  <td className="px-4 py-3">
                    <span className="font-medium text-white">
                      {c.Names?.[0]?.replace(/^\//, '') ?? c.ID.slice(0, 12)}
                    </span>
                  </td>
                  <td className="px-4 py-3 text-zinc-400">{c.Image}</td>
                  <td className="px-4 py-3">
                    <span
                      className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium ${
                        isRunning(c)
                          ? 'bg-emerald-500/20 text-emerald-400'
                          : 'bg-red-500/20 text-red-400'
                      }`}
                    >
                      <span
                        className={`w-1.5 h-1.5 rounded-full ${
                          isRunning(c) ? 'bg-emerald-400' : 'bg-red-400'
                        }`}
                      />
                      {c.State}
                    </span>
                  </td>
                  <td className="px-4 py-3">
                    <button
                      type="button"
                      onClick={() => setSelectedId(c.ID)}
                      className="inline-flex items-center gap-2 px-3 py-1.5 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-zinc-200 text-sm font-medium transition-colors"
                    >
                      <Eye className="w-4 h-4" />
                      Ver Detalhes
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {selectedId && (
        <ContainerDetailsModal
          containerId={selectedId}
          onClose={() => setSelectedId(null)}
        />
      )}
    </div>
  );
}

import { useCallback, useEffect, useMemo, useState } from 'react';
import { Eye, Loader2, Play, Square, Search, Container, Cpu, HardDrive, Layers } from 'lucide-react';
import {
  PieChart,
  Pie,
  Cell,
  ResponsiveContainer,
  Tooltip,
  Legend,
} from 'recharts';
import { api, type SystemSummary } from '../services/api';
import type { Container as ContainerType } from '../types/docker';
import { formatBytes } from '../utils/formatBytes';
import { ContainerDetailsModal } from '../components/ContainerDetailsModal';
import { toast } from 'sonner';

const SUMMARY_POLL_MS = 5000;
const SPARKLINE_POINTS = 10;

export function DashboardPage() {
  const [summary, setSummary] = useState<SystemSummary | null>(null);
  const [containers, setContainers] = useState<ContainerType[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState('');
  const [selectedId, setSelectedId] = useState<string | null>(null);
  const [actionLoading, setActionLoading] = useState<string | null>(null);
  const [cpuHistory, setCpuHistory] = useState<Record<string, number[]>>({});

  const fetchSummary = useCallback(() => {
    api.getSystemSummary().then(setSummary).catch(() => {});
  }, []);

  const fetchContainers = useCallback(() => {
    api.getContainers(true).then(setContainers).catch(() => {});
  }, []);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    Promise.all([api.getSystemSummary(), api.getContainers(true)])
      .then(([s, c]) => {
        if (!cancelled) {
          setSummary(s);
          setContainers(c);
        }
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

  useEffect(() => {
    const id = setInterval(fetchSummary, SUMMARY_POLL_MS);
    return () => clearInterval(id);
  }, [fetchSummary]);

  useEffect(() => {
    if (!summary?.container_metrics) return;
    setCpuHistory((prev) => {
      const next: Record<string, number[]> = { ...prev };
      for (const m of summary.container_metrics) {
        const arr = [...(next[m.id] ?? []), m.cpu_percentage].slice(-SPARKLINE_POINTS);
        next[m.id] = arr;
      }
      return next;
    });
  }, [summary?.container_metrics]);

  const filteredContainers = useMemo(() => {
    const q = search.trim().toLowerCase();
    if (!q) return containers;
    return containers.filter(
      (c) =>
        (c.Names?.[0] ?? '').toLowerCase().includes(q) ||
        (c.Image ?? '').toLowerCase().includes(q) ||
        (c.ID ?? '').toLowerCase().includes(q)
    );
  }, [containers, search]);

  const metricsById = useMemo(() => {
    const map: Record<string, SystemSummary['container_metrics'][0]> = {};
    for (const m of summary?.container_metrics ?? []) {
      map[m.id] = m;
    }
    return map;
  }, [summary?.container_metrics]);

  const runAction = async (containerId: string, action: 'start' | 'stop') => {
    if (actionLoading) return;
    setActionLoading(containerId);
    try {
      await api.containerAction(containerId, action);
      toast.success(`Ação "${action}" concluída`);
      fetchSummary();
      fetchContainers();
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Ação falhou');
    } finally {
      setActionLoading(null);
    }
  };

  const isRunning = (c: ContainerType) => (c.State ?? '').toLowerCase() === 'running';
  const canStart = (c: ContainerType) => !isRunning(c);
  const canStop = (c: ContainerType) => isRunning(c);

  const pieData = useMemo(() => {
    const top = summary?.top_containers_by_memory ?? [];
    if (top.length === 0) return [];
    const total = top.reduce((s, t) => s + t.memory_usage, 0);
    return top.map((t) => ({
      name: t.name || t.id.slice(0, 12),
      value: t.memory_usage,
      percent: total > 0 ? ((t.memory_usage / total) * 100).toFixed(1) : '0',
    }));
  }, [summary?.top_containers_by_memory]);

  const COLORS = ['#22c55e', '#3b82f6', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#84cc16'];

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[50vh]">
        <Loader2 className="w-10 h-10 animate-spin text-zinc-400" />
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
    <div className="p-6 space-y-6">
      <div>
        <h1 className="text-2xl font-semibold text-white">Dashboard</h1>
        <p className="text-zinc-400 text-sm mt-1">Visão geral do Docker Host</p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
        <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5 flex items-start gap-4">
          <div className="p-2.5 rounded-lg bg-emerald-500/20">
            <Container className="w-6 h-6 text-emerald-400" />
          </div>
          <div>
            <p className="text-xs font-medium text-zinc-500 uppercase tracking-wider">Containers</p>
            <p className="text-2xl font-semibold text-white mt-0.5">
              {summary?.containers_running ?? 0}
              <span className="text-zinc-500 font-normal text-base ml-1">/ {summary?.containers_total ?? 0}</span>
            </p>
            <p className="text-xs text-zinc-400 mt-1">
              {summary?.containers_stopped ?? 0} parados
            </p>
          </div>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5 flex items-start gap-4">
          <div className="p-2.5 rounded-lg bg-blue-500/20">
            <Cpu className="w-6 h-6 text-blue-400" />
          </div>
          <div>
            <p className="text-xs font-medium text-zinc-500 uppercase tracking-wider">CPU total</p>
            <p className="text-2xl font-semibold text-white mt-0.5">
              {(summary?.cpu_percent_total ?? 0).toFixed(1)}%
            </p>
            <p className="text-xs text-zinc-400 mt-1">soma dos containers ativos</p>
          </div>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5 flex items-start gap-4">
          <div className="p-2.5 rounded-lg bg-amber-500/20">
            <HardDrive className="w-6 h-6 text-amber-400" />
          </div>
          <div>
            <p className="text-xs font-medium text-zinc-500 uppercase tracking-wider">Memória</p>
            <p className="text-2xl font-semibold text-white mt-0.5">
              {formatBytes(summary?.memory_usage_bytes ?? 0)}
            </p>
            <p className="text-xs text-zinc-400 mt-1">
              de {formatBytes(summary?.memory_limit_bytes ?? 0)} total
            </p>
          </div>
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/80 p-5 flex items-start gap-4">
          <div className="p-2.5 rounded-lg bg-violet-500/20">
            <Layers className="w-6 h-6 text-violet-400" />
          </div>
          <div>
            <p className="text-xs font-medium text-zinc-500 uppercase tracking-wider">Imagens & Volumes</p>
            <p className="text-2xl font-semibold text-white mt-0.5">
              {summary?.images_count ?? 0} / {summary?.volumes_count ?? 0}
            </p>
            <p className="text-xs text-zinc-400 mt-1">imagens · volumes</p>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2 rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden">
          <div className="px-4 py-3 border-b border-zinc-800 flex items-center gap-3">
            <div className="relative flex-1">
              <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-zinc-500" />
              <input
                type="text"
                placeholder="Filtrar por nome ou imagem..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="w-full pl-9 pr-4 py-2 rounded-lg bg-zinc-800 border border-zinc-700 text-white placeholder-zinc-500 text-sm focus:outline-none focus:ring-2 focus:ring-emerald-500/50"
              />
            </div>
          </div>
          <div className="overflow-x-auto max-h-[400px] overflow-y-auto">
            <table className="w-full text-left">
              <thead className="sticky top-0 bg-zinc-900/95 border-b border-zinc-800">
                <tr>
                  <th className="px-4 py-2.5 text-xs font-medium text-zinc-500 uppercase">Status</th>
                  <th className="px-4 py-2.5 text-xs font-medium text-zinc-500 uppercase">Nome</th>
                  <th className="px-4 py-2.5 text-xs font-medium text-zinc-500 uppercase">CPU</th>
                  <th className="px-4 py-2.5 text-xs font-medium text-zinc-500 uppercase w-28">Ações</th>
                </tr>
              </thead>
              <tbody>
                {filteredContainers.map((c) => {
                  const m = metricsById[c.ID];
                  const history = cpuHistory[c.ID] ?? [];
                  return (
                    <tr
                      key={c.ID}
                      className="border-b border-zinc-800/80 hover:bg-zinc-800/30 transition-colors"
                    >
                      <td className="px-4 py-2.5">
                        <span
                          className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-xs font-medium ${
                            isRunning(c)
                              ? 'bg-emerald-500/20 text-emerald-400'
                              : 'bg-zinc-600/50 text-zinc-400'
                          }`}
                        >
                          <span
                            className={`w-1.5 h-1.5 rounded-full ${
                              isRunning(c) ? 'bg-emerald-400 animate-pulse' : 'bg-zinc-500'
                            }`}
                          />
                          {c.State}
                        </span>
                      </td>
                      <td className="px-4 py-2.5">
                        <span className="font-medium text-white">
                          {c.Names?.[0]?.replace(/^\//, '') ?? c.ID.slice(0, 12)}
                        </span>
                        <span className="text-zinc-500 text-xs ml-1 block">{c.Image}</span>
                      </td>
                      <td className="px-4 py-2.5">
                        {m ? (
                          <span className="text-zinc-300 text-sm">
                            {(m.cpu_percentage ?? 0).toFixed(1)}%
                            {history.length >= 2 && (
                              <span
                                className="ml-2 inline-flex items-end gap-0.5 h-4"
                                style={{ minWidth: `${history.length * 4}px` }}
                                title={history.map((v) => `${v.toFixed(1)}%`).join(' → ')}
                              >
                                {history.map((v, i) => (
                                  <span
                                    key={i}
                                    className="w-1 rounded-sm bg-emerald-500/70 flex-shrink-0"
                                    style={{
                                      height: `${Math.min(16, 2 + (v / 100) * 14)}px`,
                                    }}
                                  />
                                ))}
                              </span>
                            )}
                          </span>
                        ) : (
                          <span className="text-zinc-500 text-sm">—</span>
                        )}
                      </td>
                      <td className="px-4 py-2.5">
                        <div className="flex items-center gap-1">
                          <button
                            type="button"
                            onClick={() => runAction(c.ID, 'start')}
                            disabled={!canStart(c) || actionLoading === c.ID}
                            title="Iniciar"
                            className="p-1.5 rounded-lg bg-zinc-700 hover:bg-emerald-600 disabled:opacity-40 disabled:cursor-not-allowed text-zinc-300 hover:text-white transition-colors"
                          >
                            {actionLoading === c.ID ? (
                              <Loader2 className="w-4 h-4 animate-spin" />
                            ) : (
                              <Play className="w-4 h-4" />
                            )}
                          </button>
                          <button
                            type="button"
                            onClick={() => runAction(c.ID, 'stop')}
                            disabled={!canStop(c) || actionLoading === c.ID}
                            title="Parar"
                            className="p-1.5 rounded-lg bg-zinc-700 hover:bg-red-600 disabled:opacity-40 disabled:cursor-not-allowed text-zinc-300 hover:text-white transition-colors"
                          >
                            <Square className="w-4 h-4" />
                          </button>
                          <button
                            type="button"
                            onClick={() => setSelectedId(c.ID)}
                            className="p-1.5 rounded-lg bg-zinc-700 hover:bg-zinc-600 text-zinc-300 hover:text-white transition-colors"
                            title="Ver detalhes"
                          >
                            <Eye className="w-4 h-4" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
          {filteredContainers.length === 0 && (
            <p className="px-4 py-8 text-center text-zinc-500 text-sm">Nenhum container encontrado.</p>
          )}
        </div>

        <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-4">
          <h3 className="text-sm font-medium text-zinc-400 mb-4">Memória por container (top)</h3>
          {pieData.length > 0 ? (
            <ResponsiveContainer width="100%" height={280}>
              <PieChart>
                <Pie
                  data={pieData}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={90}
                  paddingAngle={2}
                  dataKey="value"
                  nameKey="name"
                >
                  {pieData.map((_, i) => (
                    <Cell key={i} fill={COLORS[i % COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip
                  formatter={(value: number | undefined) => (value != null ? formatBytes(value) : '')}
                  contentStyle={{
                    backgroundColor: '#27272a',
                    border: '1px solid #3f3f46',
                    borderRadius: '8px',
                  }}
                  labelStyle={{ color: '#a1a1aa' }}
                />
                <Legend />
              </PieChart>
            </ResponsiveContainer>
          ) : (
            <div className="h-[280px] flex items-center justify-center text-zinc-500 text-sm">
              Nenhum container em execução
            </div>
          )}
        </div>
      </div>

      {selectedId && (
        <ContainerDetailsModal containerId={selectedId} onClose={() => setSelectedId(null)} />
      )}
    </div>
  );
}

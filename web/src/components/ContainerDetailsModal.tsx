import { useCallback, useEffect, useRef, useState } from 'react';
import { X, Play, Square, RefreshCcw, Pause, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import { useDockerStats } from '../hooks/useDockerStats';
import { useDockerLogs } from '../hooks/useDockerLogs';
import { api, type ContainerAction } from '../services/api';
import type { Container } from '../types/docker';
import { formatBytes } from '../utils/formatBytes';

const stateRunning = (s: string | undefined) => (s ?? '').toLowerCase() === 'running';
const statePaused = (s: string | undefined) => (s ?? '').toLowerCase() === 'paused';
const canStart = (s: string | undefined) => {
  const lower = (s ?? '').toLowerCase();
  return lower !== 'running' && lower !== 'paused';
};

function ActionBtn({
  icon: Icon,
  label,
  onClick,
  disabled,
}: {
  icon: React.ComponentType<{ className?: string }>;
  label: string;
  onClick: () => void;
  disabled: boolean;
}) {
  return (
    <button
      type="button"
      title={label}
      onClick={onClick}
      disabled={disabled}
      className="flex items-center justify-center w-9 h-9 rounded-lg border border-zinc-600 bg-zinc-800/80 text-zinc-300 hover:bg-zinc-700 hover:text-white disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
    >
      <Icon className="w-4 h-4" />
    </button>
  );
}

interface ContainerDetailsModalProps {
  containerId: string;
  onClose: () => void;
}

export function ContainerDetailsModal({
  containerId,
  onClose,
}: ContainerDetailsModalProps) {
  const [container, setContainer] = useState<Container | null>(null);
  const [actionLoading, setActionLoading] = useState(false);
  const { metrics, connected, error: wsError } = useDockerStats(containerId);
  const { text: logsText, connected: logsConnected, error: logsError } = useDockerLogs(
    containerId,
    connected
  );
  const logsEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logsText]);

  const refreshContainer = useCallback(() => {
    api
      .getContainers(true)
      .then((list) => {
        const c = list.find((x) => x.ID === containerId || x.ID.startsWith(containerId));
        if (c) setContainer(c);
      })
      .catch(() => {});
  }, [containerId]);

  useEffect(() => {
    refreshContainer();
  }, [refreshContainer]);

  const runAction = async (action: ContainerAction) => {
    if (actionLoading) return;
    setActionLoading(true);
    try {
      await api.containerAction(containerId, action);
      toast.success(`Ação "${action}" concluída`);
      refreshContainer();
      setTimeout(() => refreshContainer(), 600);
    } catch (e) {
      toast.error(e instanceof Error ? e.message : 'Ação falhou');
    } finally {
      setActionLoading(false);
    }
  };

  const chartData = metrics.map((m) => ({
    time: new Date(m.timestamp ?? 0).toLocaleTimeString(),
    cpu: m.cpu_percentage ?? 0,
    memory: m.memory_percent ?? 0,
    memoryUsage: m.memory_usage ?? 0,
    memoryLimit: m.memory_limit ?? 0,
  }));

  const latest = metrics[metrics.length - 1];

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        className="bg-zinc-900 border border-zinc-700 rounded-xl shadow-2xl w-full max-w-4xl max-h-[90vh] overflow-hidden flex flex-col"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-center justify-between px-6 py-4 border-b border-zinc-800">
          <div>
            <h2 className="text-lg font-semibold text-white">
              {container?.Names?.[0]?.replace(/^\//, '') ?? containerId.slice(0, 12)}
            </h2>
            <p className="text-sm text-zinc-400 mt-0.5">
              {container?.Image ?? '-'} · {container?.State ?? '-'}
            </p>
          </div>
          <button
            type="button"
            onClick={onClose}
            className="p-2 rounded-lg hover:bg-zinc-800 text-zinc-400 hover:text-white transition-colors"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="px-6 py-4 border-b border-zinc-800 flex flex-wrap gap-4">
          <div className="text-sm">
            <span className="text-zinc-500">ID:</span>{' '}
            <span className="text-zinc-300 font-mono text-xs">
              {container?.ID?.slice(0, 12) ?? containerId.slice(0, 12)}
            </span>
          </div>
          <div className="text-sm">
            <span className="text-zinc-500">Status:</span>{' '}
            <span className="text-zinc-300">{container?.Status ?? '-'}</span>
          </div>
          {connected && (
            <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-emerald-500/20 text-emerald-400">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
              Live
            </span>
          )}
          {wsError && (
            <span className="text-red-400 text-sm">{wsError}</span>
          )}
        </div>

        {latest && (
          <div className="px-6 py-3 border-b border-zinc-800 grid grid-cols-2 sm:grid-cols-4 gap-4">
            <div>
              <p className="text-xs text-zinc-500">CPU</p>
              <p className="text-lg font-semibold text-white">
                {(latest.cpu_percentage ?? 0).toFixed(2)}%
              </p>
            </div>
            <div>
              <p className="text-xs text-zinc-500">Memória</p>
              <p className="text-lg font-semibold text-white">
                {(latest.memory_percent ?? 0).toFixed(2)}%
              </p>
            </div>
            <div>
              <p className="text-xs text-zinc-500">Uso</p>
              <p className="text-sm font-medium text-zinc-300">
                {formatBytes(latest.memory_usage ?? 0)}
              </p>
            </div>
            <div>
              <p className="text-xs text-zinc-500">Limite</p>
              <p className="text-sm font-medium text-zinc-300">
                {formatBytes(latest.memory_limit ?? 0)}
              </p>
            </div>
          </div>
        )}

        <div className="flex-1 min-h-0 p-6">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-sm text-zinc-400 mr-2">Ações</span>
            {(container?.State ?? '').toLowerCase() === 'exited' && (
              <span className="text-xs text-amber-400/90">Container parado — use Iniciar para subir novamente.</span>
            )}
            {actionLoading && (
              <Loader2 className="w-4 h-4 text-amber-400 animate-spin" aria-hidden />
            )}
            <div className="flex flex-wrap gap-1.5">
              <ActionBtn
                icon={Play}
                label="Iniciar"
                onClick={() => runAction('start')}
                disabled={actionLoading || !canStart(container?.State)}
              />
              <ActionBtn
                icon={Square}
                label="Parar"
                onClick={() => runAction('stop')}
                disabled={actionLoading || (!stateRunning(container?.State) && !statePaused(container?.State))}
              />
              <ActionBtn
                icon={RefreshCcw}
                label="Reiniciar"
                onClick={() => runAction('restart')}
                disabled={actionLoading || container?.State !== 'running'}
              />
              <ActionBtn
                icon={Pause}
                label="Pausar"
                onClick={() => runAction('pause')}
                disabled={actionLoading || container?.State !== 'running'}
              />
              <ActionBtn
                icon={Play}
                label="Retomar"
                onClick={() => runAction('unpause')}
                disabled={actionLoading || container?.State !== 'paused'}
              />
            </div>
          </div>
          <p className="text-sm text-zinc-400 mb-2">Métricas em tempo real</p>
          <div className="w-full min-w-0" style={{ height: 256 }}>
            {chartData.length > 0 ? (
              <ResponsiveContainer width="100%" height={256}>
                <LineChart data={chartData} margin={{ top: 5, right: 20, left: 0, bottom: 5 }}>
                  <CartesianGrid strokeDasharray="3 3" stroke="#27272a" />
                  <XAxis
                    dataKey="time"
                    stroke="#71717a"
                    tick={{ fontSize: 11 }}
                  />
                  <YAxis
                    yAxisId="percent"
                    stroke="#71717a"
                    tick={{ fontSize: 11 }}
                    tickFormatter={(v) => `${v}%`}
                    domain={[0, 100]}
                  />
                  <Tooltip
                    contentStyle={{
                      backgroundColor: '#27272a',
                      border: '1px solid #3f3f46',
                      borderRadius: '8px',
                    }}
                    labelStyle={{ color: '#a1a1aa' }}
                    formatter={(value, name) => [
                      typeof value === 'number' ? `${value.toFixed(2)}%` : '',
                      name === 'cpu' ? 'CPU' : 'Memória',
                    ]}
                  />
                  <Legend />
                  <Line
                    yAxisId="percent"
                    type="monotone"
                    dataKey="cpu"
                    name="CPU"
                    stroke="#22c55e"
                    strokeWidth={2}
                    dot={false}
                  />
                  <Line
                    yAxisId="percent"
                    type="monotone"
                    dataKey="memory"
                    name="Memória"
                    stroke="#3b82f6"
                    strokeWidth={2}
                    dot={false}
                  />
                </LineChart>
              </ResponsiveContainer>
            ) : (
              <div
                className="flex items-center justify-center rounded-lg bg-zinc-800/50 border border-zinc-700 text-zinc-500 text-sm"
                style={{ height: 256 }}
              >
                {connected
                  ? 'Aguardando métricas...'
                  : 'Conectando ao stream...'}
              </div>
            )}
          </div>

          <div className="mt-4">
            <p className="text-sm text-zinc-400 mb-2 flex items-center gap-2">
              Logs
              {logsConnected && (
                <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-xs font-medium bg-emerald-500/20 text-emerald-400">
                  <span className="w-1 h-1 rounded-full bg-emerald-400 animate-pulse" />
                  stream
                </span>
              )}
              {logsError && <span className="text-red-400 text-xs">{logsError}</span>}
            </p>
            <div
              className="w-full rounded-lg bg-zinc-950 border border-zinc-700 overflow-hidden font-mono text-xs text-zinc-300"
              style={{ height: 200 }}
            >
              <pre className="h-full overflow-auto p-3 whitespace-pre-wrap break-all m-0">
                {logsText || (logsConnected ? 'Aguardando logs...' : 'Conectando aos logs...')}
                <span ref={logsEndRef} />
              </pre>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

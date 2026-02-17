import { useEffect, useState } from 'react';
import { Loader2 } from 'lucide-react';
import { api } from '../services/api';
import type { Image } from '../types/docker';
import { formatBytes } from '../utils/formatBytes';

export function ImagesPage() {
  const [images, setImages] = useState<Image[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    api
      .getImages()
      .then((data) => {
        if (!cancelled) setImages(data);
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
        <h1 className="text-2xl font-semibold text-white">Images</h1>
        <p className="text-zinc-400 text-sm mt-1">{images.length} imagem(ns)</p>
      </div>

      <div className="rounded-xl border border-zinc-800 bg-zinc-900/50 overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-left">
            <thead>
              <tr className="border-b border-zinc-800">
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Repo / Tag
                </th>
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  ID
                </th>
                <th className="px-4 py-3 text-xs font-medium text-zinc-500 uppercase tracking-wider">
                  Tamanho
                </th>
              </tr>
            </thead>
            <tbody>
              {images.map((img) => (
                <tr
                  key={img.ID}
                  className="border-b border-zinc-800/80 hover:bg-zinc-800/30 transition-colors"
                >
                  <td className="px-4 py-3 text-white font-medium">
                    {img.RepoTags?.[0] ?? '<none>'}
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-zinc-400">
                    {img.ID.replace('sha256:', '').slice(0, 12)}
                  </td>
                  <td className="px-4 py-3 text-zinc-400">
                    {formatBytes(img.Size)}
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

/*
   note : badge skor prioritas (skala 0–100), dipakai di Jadwal Produksi dan detail pesanan
*/
export default function ScoreBadge({ score }) {
  if (score == null) return <span className="text-[#8a7768] text-xs">—</span>;
  const s = Math.round(score * 100);
  const color = s >= 85 ? 'bg-[#eaf5ed] text-[#327a52]'
              : s >= 70 ? 'bg-[#fdf6ec] text-[#a66a17]'
              :           'bg-[#fef2f2] text-[#b74b40]';
  return <span className={`inline-flex items-center px-2 py-0.5 rounded-lg text-xs font-medium ${color}`}>{s}</span>;
}

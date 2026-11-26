"use client";

import { useEffect, useState } from "react";
import styles from "./backup-story.module.css";

type V = [number, number];

// Isometric projection: +x runs right-down, +y runs left-down, +z is up.
const P = (x: number, y: number, z = 0): V => [
  Math.round((x - y) * 0.866 * 10) / 10,
  Math.round(((x + y) * 0.5 - z) * 10) / 10,
];
const pts = (...v: V[]) => v.map((p) => p.join(",")).join(" ");
const line = (...v: V[]) => "M" + v.map((p) => p.join(" ")).join("L");

type Part = { x: number; y: number; z?: number; w: number; d: number; h: number };

function Box({ x, y, z = 0, w, d, h }: Part) {
  const t = z + h;
  return (
    <>
      <polygon className={styles.left} points={pts(P(x, y + d, z), P(x + w, y + d, z), P(x + w, y + d, t), P(x, y + d, t))} />
      <polygon className={styles.right} points={pts(P(x + w, y, z), P(x + w, y + d, z), P(x + w, y + d, t), P(x + w, y, t))} />
      <polygon className={styles.top} points={pts(P(x, y, t), P(x + w, y, t), P(x + w, y + d, t), P(x, y + d, t))} />
    </>
  );
}

function Cylinder({ x, y, z, r, h, bands = 0 }: { x: number; y: number; z: number; r: number; h: number; bands?: number }) {
  const [bx, by] = P(x, y, z);
  const ty = by - h;
  const rx = r * 1.2247;
  const ry = r * 0.7071;
  const arcs = Array.from({ length: bands }, (_, i) => by - (h * (i + 1)) / (bands + 1));
  const side = "M" + (bx - rx) + " " + ty + "L" + (bx - rx) + " " + by + "A" + rx + " " + ry + " 0 0 0 " + (bx + rx) + " " + by + "L" + (bx + rx) + " " + ty + "Z";
  return (
    <>
      <path className={styles.left} d={side} />
      {arcs.map((ay) => (
        <path key={ay} className={styles.seam} d={"M" + (bx - rx) + " " + ay + "A" + rx + " " + ry + " 0 0 0 " + (bx + rx) + " " + ay} />
      ))}
      <ellipse className={styles.top} cx={bx} cy={ty} rx={rx} ry={ry} />
    </>
  );
}

function Halo({ x, y, w, d }: { x: number; y: number; w: number; d: number }) {
  return <polygon className={styles.halo} points={pts(P(x, y), P(x + w, y), P(x + w, y + d), P(x, y + d))} />;
}

function Label({ at, text }: { at: V; text: string }) {
  const width = text.length * 7 + 22;
  return (
    <g className={styles.label} transform={"translate(" + at[0] + " " + at[1] + ")"}>
      <rect x={-width / 2} y={-13} width={width} height={26} rx={7} />
      <text x={0} y={4.5} textAnchor="middle">{text}</text>
    </g>
  );
}

// Data paths, in world coordinates, follow the code's order of operations.
const routes = {
  connect: line(P(75, 335, 41), P(75, 205, 41)),
  dump: line(P(90, 170, 14), P(130, 170, 14), P(130, 130, 14)),
  write: line(P(150, 110, 14), P(192, 110, 14), P(192, 85, 14)),
  upload: line(P(192, 55, 44), P(192, 55, 172), P(192, -46, 172)),
  verify: line(P(226, 55, 14), P(290, 55, 14), P(290, 110, 14), P(338, 110, 14)),
};
const packets: Record<number, string> = {
  0: routes.connect,
  1: routes.dump + routes.write.replace("M", "L"),
  2: routes.upload,
  3: routes.verify,
};

const usage: Record<string, number[]> = {
  database: [0],
  tunnel: [0],
  tool: [1],
  gzip: [1],
  folder: [1, 2, 3],
  bucket: [2],
  container: [3],
};

export const parts: Record<string, { name: string; at: V }> = {
  database: { name: "Database", at: [P(75, 365, 66)[0], P(75, 365, 66)[1] - 30] },
  tunnel: { name: "SSH tunnel", at: [-112, 178] },
  tool: { name: "pg_dump", at: [P(60, 170, 52)[0], P(60, 170, 52)[1] - 34] },
  gzip: { name: "gzip", at: [P(130, 110, 38)[0], P(130, 110, 38)[1] - 34] },
  folder: { name: "backups/<connection>/", at: [P(193, 55, 36)[0] + 30, P(193, 55, 36)[1] + 66] },
  bucket: { name: "S3 bucket", at: [P(192, -80, 196)[0], P(192, -80, 196)[1] - 30] },
  container: { name: "Throwaway database", at: [150, 290] },
};
export const partsIn = (stage: number) => Object.keys(usage).filter((k) => usage[k].includes(stage));

function useReducedMotion() {
  const [reduced, setReduced] = useState(true);
  useEffect(() => {
    const query = window.matchMedia("(prefers-reduced-motion: reduce)");
    const update = () => setReduced(query.matches);
    update();
    query.addEventListener("change", update);
    return () => query.removeEventListener("change", update);
  }, []);
  return reduced;
}

export function BackupFigure({ active, labels = true }: { active: number; labels?: boolean }) {
  const reduced = useReducedMotion();
  const state = (name: string) => {
    const stages = usage[name];
    if (stages.includes(active)) return "on";
    return stages[0] < active ? "past" : "off";
  };
  const flow = (stage: number) => (active === stage ? "on" : stage < active ? "past" : "off");

  return (
    <svg className={styles.svg} viewBox="-340 -210 665 520" data-stage={active} aria-hidden="true">
      <Halo x={-6} y={-6} w={252} d={232} />
      <Halo x={26} y={316} w={98} d={98} />
      <Halo x={316} y={56} w={108} d={108} />

      <g className={styles.part} data-state={state("bucket")}>
        <ellipse className={styles.halo} cx={P(192, -80, 150)[0]} cy={P(192, -80, 150)[1] + 8} rx={46} ry={16} />
        <Cylinder x={192} y={-80} z={150} r={34} h={46} bands={1} />
      </g>

      <g className={styles.base}>
        <Box x={0} y={0} w={240} d={220} h={12} />
        <path className={styles.seam} d={line(P(12, 12, 12), P(228, 12, 12), P(228, 208, 12))} />
      </g>

      <path className={styles.flow} data-state={flow(1)} d={routes.dump} />
      <path className={styles.flow} data-state={flow(1)} d={routes.write} />
      <path className={styles.flow} data-state={flow(3)} d={routes.verify} />

      <g className={styles.part} data-state={state("folder")}>
        <Box x={160} y={25} z={12} w={66} d={60} h={8} />
        <Box x={170} y={35} z={20} w={46} d={40} h={4} />
        <Box x={170} y={35} z={26} w={46} d={40} h={4} />
        <g className={styles.file} data-purge={active === 2 ? "on" : "off"}>
          <Box x={170} y={35} z={32} w={46} d={40} h={4} />
        </g>
      </g>

      <g className={styles.part} data-state={state("gzip")}>
        <Box x={110} y={90} z={12} w={40} d={40} h={26} />
        <path className={styles.seam} d={line(P(118, 130, 30), P(142, 130, 30))} />
        <path className={styles.seam} d={line(P(118, 130, 24), P(136, 130, 24))} />
      </g>

      <g className={styles.part} data-state={state("tool")}>
        <Box x={30} y={140} z={12} w={60} d={60} h={40} />
        <path className={styles.seam} d={line(P(40, 200, 40), P(52, 200, 36), P(40, 200, 32))} />
        <path className={styles.seam} d={line(P(56, 200, 32), P(76, 200, 32))} />
      </g>

      <g className={styles.part} data-state={state("tunnel")}>
        <Box x={68} y={200} z={27} w={14} d={135} h={14} />
      </g>
      <path className={styles.flow} data-state={flow(0)} d={routes.connect} />

      <g className={styles.part} data-state={state("database")}>
        <Box x={30} y={320} w={90} d={90} h={10} />
        <Cylinder x={75} y={365} z={10} r={30} h={56} bands={2} />
      </g>

      <g className={styles.part} data-state={state("container")} data-tone="green">
        <Box x={320} y={60} w={100} d={100} h={8} />
        <g className={styles.ephemeral}>
          <Box x={335} y={75} z={8} w={70} d={70} h={50} />
        </g>
        {[90, 105, 120, 135].map((y) => (
          <path key={y} className={styles.seam} d={line(P(405, y, 16), P(405, y, 50))} />
        ))}
        <Cylinder x={370} y={110} z={58} r={16} h={20} bands={1} />
      </g>

      <path className={styles.flow} data-state={flow(2)} d={routes.upload} />

      {!reduced && packets[active] !== undefined &&
        [0, 1, 2].map((i) => (
          <circle key={active + "-" + i} className={styles.packet} r={3}>
            <animateMotion dur="2.4s" begin={i * 0.8 + "s"} repeatCount="indefinite" path={packets[active]} />
          </circle>
        ))}

      {labels && partsIn(active).map((key) => <Label key={key} at={parts[key].at} text={parts[key].name} />)}
    </svg>
  );
}


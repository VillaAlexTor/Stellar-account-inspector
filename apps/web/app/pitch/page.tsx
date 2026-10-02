import type { Metadata } from "next";
import { PitchDeck } from "@/components/pitch/PitchDeck";

export const metadata: Metadata = {
  title: "Pitch | Stellar Account Inspector",
  description: "Presentación interactiva de Stellar Account Inspector.",
};

export default function PitchPage() {
  return <PitchDeck />;
}

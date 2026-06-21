import { Navbar } from "../../components/landing/Navbar";

import { FeaturesSection } from "../../components/landing/FeaturesSection";
import { HowItWorks } from "../../components/landing/HowItWorks";
import { CTASection } from "../../components/landing/CTASection";
import { Footer } from "../../components/landing/Footer";
import { HeroSection } from "@/components/landing/HeroSection";
import TeamSection from "@/components/landing/TeamSection";

export default function LandingPage() {
  return (
    <div className="landing-page min-h-screen  text-gray-900 font-sans selection:bg-emerald-500/30">
      <Navbar />
      <main>
        <HeroSection />
        <FeaturesSection />
        <TeamSection/>
        <HowItWorks />
        <CTASection />
      </main>
      <Footer />
    </div>
  );
}

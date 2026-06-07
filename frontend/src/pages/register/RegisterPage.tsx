import LeftSide from "@/components/login/LeftSide";
import RegisterSide from "@/components/login/RegisterSide";


export default function RegisterPage() {
    return (
      <section className="relative flex min-h-screen w-full items-stretch justify-center overflow-x-hidden">
        <LeftSide />
        <RegisterSide />
      </section>
    )
  }

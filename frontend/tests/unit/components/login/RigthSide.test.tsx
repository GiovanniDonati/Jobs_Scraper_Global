/* eslint-disable @typescript-eslint/no-explicit-any */
import { fireEvent, render, screen } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";

vi.mock("@unpic/react", () => ({
  Image: (props: any) => <img {...props} alt={props.alt} />,
}));

vi.mock("framer-motion", () => ({
  motion: {
    div: ({ children, ...props }: any) => <div {...props}>{children}</div>,
    button: ({ children, ...props }: any) => <button {...props}>{children}</button>,
  },
}));

import RigthSide from "@/components/login/RigthSide";

describe("RigthSide", () => {
  const logSpy = vi.spyOn(console, "log").mockImplementation(() => {});

  beforeEach(() => {
    logSpy.mockClear();
  });

  it("renderiza formulário e alterna visibilidade da senha", () => {
    render(<RigthSide />);

    const passwordInput = screen.getByLabelText(/senha/i) as HTMLInputElement;
    expect(passwordInput.type).toBe("password");

    const revealButtons = screen.getAllByRole("button", { name: "" });
    fireEvent.click(revealButtons[0]);

    expect((screen.getByLabelText(/senha/i) as HTMLInputElement).type).toBe("text");
  });

  it("mostra erros para submit invalido", () => {
    render(<RigthSide />);

    fireEvent.click(screen.getByRole("button", { name: /entrar/i }));

    expect(screen.getByText(/campo de e-mail.*obrigatório/i)).toBeInTheDocument();
    expect(screen.getByText(/campo de senha.*obrigatório/i)).toBeInTheDocument();
  });

  it("valida formato de email e tamanho minimo da senha", () => {
    render(<RigthSide />);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "qa@teste" } });
    fireEvent.change(screen.getByLabelText(/senha/i), { target: { value: "123" } });
    fireEvent.click(screen.getByRole("button", { name: /entrar/i }));

    expect(screen.getByText(/insira um e-mail válido/i)).toBeInTheDocument();
    expect(screen.getByText(/pelo menos 6 caracteres/i)).toBeInTheDocument();
  });

  it("envia formulário valido", () => {
    render(<RigthSide />);

    fireEvent.change(screen.getByLabelText(/email/i), { target: { value: "qa@teste.com" } });
    fireEvent.change(screen.getByLabelText(/senha/i), { target: { value: "123456" } });
    fireEvent.click(screen.getByRole("button", { name: /entrar/i }));

    expect(logSpy).toHaveBeenCalledWith("Formulário válido!", {
      email: "qa@teste.com",
      password: "123456",
    });
  });
});

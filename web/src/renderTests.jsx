import React from "react";
import { render } from "@testing-library/react";
import { MemoryRouter } from "react-router-dom";

const Wrapper = ({ children }) => {
  return <MemoryRouter>{children}</MemoryRouter>;
};

const renderWithWrapper = (ui, options) => {
  return render(ui, { wrapper: Wrapper, ...options });
};

// Re-export everything.
export * from "@testing-library/react";

// Override the render method.
export { renderWithWrapper as render };

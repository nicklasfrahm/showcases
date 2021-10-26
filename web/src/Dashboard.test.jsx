import { render, screen } from "./renderTests";
import Dashboard from "./Dashboard";

it("displays the Email Sender", async () => {
  render(<Dashboard />);

  expect(screen.getByRole("heading")).toHaveTextContent("Email Sender");
});

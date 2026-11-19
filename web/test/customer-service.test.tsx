import { expect, test } from "bun:test";
import { renderToStaticMarkup } from "react-dom/server";
import { CustomerServiceContent } from "../src/components/layout/customer-service-content";
import { DEFAULT_CUSTOMER_SERVICE, validateCustomerService } from "../src/lib/customer-service";

test("second contact alone can enable customer service and is validated", () => {
    const value = { ...DEFAULT_CUSTOMER_SERVICE, enabled: true, secondWechat: "support-two" };
    expect(validateCustomerService(value)).toBe("");
    expect(validateCustomerService({ ...value, secondWechat: "" })).not.toBe("");
    expect(validateCustomerService({ ...value, secondContactName: "字".repeat(81) })).not.toBe("");
});

test("contact renderer keeps each QR and contact details together", () => {
    const html = renderToStaticMarkup(
        <CustomerServiceContent value={{ ...DEFAULT_CUSTOMER_SERVICE, enabled: true, contactName: "客服甲", phone: "111", qrCodeUrl: "/first.png", secondContactName: "客服乙", secondWechat: "wechat-two", secondQrCodeUrl: "/second.png" }} />,
    );
    const cards = html.match(/<section\b[^>]*>[\s\S]*?<\/section>/g) || [];
    expect(cards).toHaveLength(2);
    expect(cards[0]).toContain("/first.png");
    expect(cards[0]).toContain("111");
    expect(cards[0]).not.toContain("wechat-two");
    expect(cards[1]).toContain("/second.png");
    expect(cards[1]).toContain("wechat-two");
});

test("empty second contact does not render an empty card", () => {
    const html = renderToStaticMarkup(<CustomerServiceContent value={{ ...DEFAULT_CUSTOMER_SERVICE, contactName: "客服甲", qrCodeUrl: "", secondQrCodeUrl: "" }} />);
    expect(html.match(/<section\b/g)).toHaveLength(1);
});

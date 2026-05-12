#import <Cocoa/Cocoa.h>

static NSStatusItem *statusItem = nil;
static id menuDelegate = nil;
static NSImage *appIconImage = nil;

// Go 导出函数
extern void trayShowWindow();
extern void trayQuitApp();

@interface TrayMenuDelegate : NSObject
- (void)showWindow:(id)sender;
- (void)quitApp:(id)sender;
@end

@implementation TrayMenuDelegate
- (void)showWindow:(id)sender {
    trayShowWindow();
}
- (void)quitApp:(id)sender {
    trayQuitApp();
}
@end

void setupTrayIcon(const char* iconPath, const char* tooltip) {
    NSString *tooltipStr = tooltip ? [NSString stringWithUTF8String:tooltip] : @"StockFinLens 财报透镜";
    NSString *iconPathStr = iconPath ? [NSString stringWithUTF8String:iconPath] : nil;

    dispatch_async(dispatch_get_main_queue(), ^{
        NSStatusBar *bar = [NSStatusBar systemStatusBar];
        if (!bar) {
            NSLog(@"[Tray] NSStatusBar is nil");
            return;
        }

        statusItem = [bar statusItemWithLength:NSVariableStatusItemLength];
        if (!statusItem) {
            NSLog(@"[Tray] NSStatusItem is nil");
            return;
        }

        statusItem.visible = YES;

        NSButton *button = statusItem.button;
        if (!button) {
            NSLog(@"[Tray] button is nil");
            return;
        }

        // 加载图标
        if (iconPathStr) {
            appIconImage = [[NSImage alloc] initWithContentsOfFile:iconPathStr];
        }

        if (!appIconImage) {
            appIconImage = [[NSImage alloc] initWithSize:NSMakeSize(18, 18)];
            [appIconImage lockFocus];
            [[NSColor colorWithDeviceRed:0.2 green:0.5 blue:0.9 alpha:1.0] set];
            NSRectFill(NSMakeRect(0, 0, 18, 18));
            [appIconImage unlockFocus];
        }

        // 保留原始颜色（非 template），以便在合成图片时显示彩色 logo
        // 不预先缩放，保留原始分辨率，在合成时再按需绘制
        appIconImage.template = NO;

        button.image = appIconImage;
        button.title = @" SFL";
        button.imagePosition = NSImageLeading;
        button.toolTip = tooltipStr;

        NSLog(@"[Tray] button configured: image=%@ title=%@ frame=%@",
              button.image ? @"YES" : @"NO",
              button.title,
              NSStringFromRect(button.frame));

        // 菜单
        menuDelegate = [[TrayMenuDelegate alloc] init];
        NSMenu *menu = [[NSMenu alloc] init];

        NSMenuItem *showItem = [[NSMenuItem alloc] initWithTitle:@"显示主窗口"
                                                          action:@selector(showWindow:)
                                                   keyEquivalent:@""];
        showItem.target = menuDelegate;
        [menu addItem:showItem];

        [menu addItem:[NSMenuItem separatorItem]];

        NSMenuItem *quitItem = [[NSMenuItem alloc] initWithTitle:@"退出"
                                                         action:@selector(quitApp:)
                                                  keyEquivalent:@"q"];
        quitItem.target = menuDelegate;
        [menu addItem:quitItem];

        statusItem.menu = menu;

        NSLog(@"[Tray] === status item ready ===");
    });
}

// renderCombinedImage 将 logo + 彩色文字合成为一张 NSImage
static NSImage* renderCombinedImage(NSString *text, NSColor *color) {
    NSFont *font = [NSFont menuBarFontOfSize:0];
    if (!font) {
        font = [NSFont systemFontOfSize:13];
    }

    NSDictionary *attrs = @{
        NSFontAttributeName: font,
        NSForegroundColorAttributeName: color
    };

    NSSize textSize = [text sizeWithAttributes:attrs];
    CGFloat textWidth = MAX(ceil(textSize.width), 1.0);
    CGFloat textHeight = MAX(ceil(textSize.height), 1.0);

    CGFloat logoSize = 22.0;
    CGFloat padding = 4.0;
    CGFloat totalWidth = logoSize + padding + textWidth;

    NSImage *image = [NSImage imageWithSize:NSMakeSize(totalWidth, 22) flipped:NO drawingHandler:^BOOL(NSRect dstRect) {
        // 绘制 logo（垂直居中）
        if (appIconImage) {
            NSRect logoRect = NSMakeRect(0, (22 - logoSize) / 2.0, logoSize, logoSize);
            [appIconImage drawInRect:logoRect];
        }

        // 绘制彩色文字（垂直居中）
        NSPoint point = NSMakePoint(logoSize + padding, (22 - textHeight) / 2.0);
        [text drawAtPoint:point withAttributes:attrs];

        return YES;
    }];

    return image;
}

void updateTrayTitle(const char* title, double changePercent) {
    // 关键：Go 端的 C.CString 会在 updateTrayTitle 返回后被 defer free 释放。
    // 由于我们内部使用了 dispatch_async，block 在之后才执行，届时 title 指针已是野指针。
    // 因此必须在同步入口处立即把 char* 复制成 NSString。
    NSString *titleStr = title ? [NSString stringWithUTF8String:title] : @"SFL";

    dispatch_async(dispatch_get_main_queue(), ^{
        if (!statusItem || !statusItem.button) {
            NSLog(@"[Tray] updateTrayTitle called but statusItem/button is nil");
            return;
        }

        NSButton *button = statusItem.button;

        // 根据涨跌设置颜色：A股涨红跌绿
        NSColor *textColor;
        if (changePercent > 0) {
            textColor = [NSColor colorWithDeviceRed:0.92 green:0.25 blue:0.20 alpha:1.0]; // 红
        } else if (changePercent < 0) {
            textColor = [NSColor colorWithDeviceRed:0.20 green:0.75 blue:0.30 alpha:1.0]; // 绿
        } else {
            textColor = [NSColor labelColor]; // 平盘用系统默认色
        }

        // 合成 logo + 彩色文字为一张图片
        NSImage *combinedImage = renderCombinedImage(titleStr, textColor);
        button.image = combinedImage;
        button.title = @"";

        NSLog(@"[Tray] updated title: %@ (change=%.2f%%)", titleStr, changePercent);
    });
}

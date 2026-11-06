// Puts each captured screen under its headline at the exact App Store size.
// It fails on a capture of the wrong size or a blank one, and screenshots.sh runs it.
import CoreGraphics
import CoreText
import Foundation
import ImageIO
import UniformTypeIdentifiers

/// Pixels across and down.
struct Pixels: Decodable {
  let width: Int
  let height: Int
}

/// A store size from the app's config: its folder, and the pixels of the final image.
struct Size: Decodable {
  let folder: String
  let width: Int
  let height: Int
  /// Whether the status bar shows today's date, which simctl can't fix, so it comes off.
  var datedStatusBar: Bool?
  /// The capture's pixels, and the top-left corner the screen fills, when they aren't the size's.
  var capture: Pixels?
  var screen: Pixels?
}

/// What this tool reads from the app's config.
struct Config: Decodable {
  /// Red, green and blue from 0 to 1, behind the headline.
  let background: [CGFloat]
  let sizes: [Size]
}

/// The share of changed pixels that two runs of the same app reach; below it a file stays.
let keepsBelow = 0.0002

func fail(_ message: String) -> Never {
  FileHandle.standardError.write(Data("compose-screenshots: \(message)\n".utf8))
  exit(1)
}

func load(_ url: URL) -> CGImage {
  guard let source = CGImageSourceCreateWithURL(url as CFURL, nil),
    let image = CGImageSourceCreateImageAtIndex(source, 0, nil)
  else { fail("no capture at \(url.path)") }
  return image
}

/// Fails unless a thumbnail of the screen holds many different shades.
func requireContent(_ image: CGImage, name: String) {
  let side = 64
  var pixels = [UInt8](repeating: 0, count: side * side)
  pixels.withUnsafeMutableBytes { bytes in
    let context = CGContext(
      data: bytes.baseAddress, width: side, height: side, bitsPerComponent: 8, bytesPerRow: side,
      space: CGColorSpaceCreateDeviceGray(), bitmapInfo: CGImageAlphaInfo.none.rawValue)
    context?.draw(image, in: CGRect(x: 0, y: 0, width: side, height: side))
  }
  let mean = pixels.reduce(0) { $0 + Double($1) } / Double(pixels.count)
  let spread = pixels.reduce(0) { $0 + abs(Double($1) - mean) } / Double(pixels.count)
  guard spread > 8 else { fail("\(name) looks blank, spread \(spread)") }
}

/// The screen a capture shows, checked against the size it must have.
func screen(in capture: CGImage, for size: Size, name: String) -> CGImage {
  let expected = size.capture ?? Pixels(width: size.width, height: size.height)
  guard capture.width == expected.width, capture.height == expected.height else {
    fail("\(name) is \(capture.width)x\(capture.height), expected \(expected)")
  }
  let corner = size.screen ?? Pixels(width: size.width, height: size.height)
  let crop = CGRect(x: 0, y: 0, width: corner.width, height: corner.height)
  guard let shown = capture.cropping(to: crop) else { fail("\(name) won't crop") }
  requireContent(shown, name: name)
  return size.datedStatusBar == true ? withoutStatusBar(shown) : shown
}

/// Paints over a dated status bar with the row of pixels just below it, so the images of two
/// days match. Its height in pixels is 24 points at 2x, plus a little.
func withoutStatusBar(_ shot: CGImage) -> CGImage {
  let bar = 52
  let width = CGFloat(shot.width)
  let height = CGFloat(shot.height)
  guard
    let context = CGContext(
      data: nil, width: shot.width, height: shot.height, bitsPerComponent: 8, bytesPerRow: 0,
      space: CGColorSpace(name: CGColorSpace.sRGB)!,
      bitmapInfo: CGImageAlphaInfo.noneSkipLast.rawValue),
    let row = shot.cropping(to: CGRect(x: 0, y: bar, width: shot.width, height: 1))
  else { fail("no canvas") }
  context.draw(shot, in: CGRect(x: 0, y: 0, width: width, height: height))
  context.draw(row, in: CGRect(x: 0, y: height - CGFloat(bar), width: width, height: CGFloat(bar)))
  guard let image = context.makeImage() else { fail("no image") }
  return image
}

/// Draws the headline, then the screen below it with rounded corners.
func compose(_ shot: CGImage, headline: String, size: Size, background: CGColor) -> CGImage {
  let width = CGFloat(size.width)
  let height = CGFloat(size.height)
  guard
    let context = CGContext(
      data: nil, width: size.width, height: size.height, bitsPerComponent: 8, bytesPerRow: 0,
      space: CGColorSpace(name: CGColorSpace.sRGB)!,
      bitmapInfo: CGImageAlphaInfo.noneSkipLast.rawValue)
  else { fail("no canvas") }
  context.setFillColor(background)
  context.fill(CGRect(x: 0, y: 0, width: width, height: height))

  let margin = width * 0.07
  let top = height * 0.21
  let fontSize = min(width * 0.085, height * 0.05)
  let font = CTFontCreateUIFontForLanguage(.system, fontSize, nil)
  let bold = font.flatMap { CTFontCreateCopyWithSymbolicTraits($0, fontSize, nil, .boldTrait, .boldTrait) }
  let words = NSAttributedString(
    string: headline,
    attributes: [
      kCTFontAttributeName as NSAttributedString.Key: bold ?? font as Any,
      kCTForegroundColorAttributeName as NSAttributedString.Key:
        CGColor(red: 1, green: 1, blue: 1, alpha: 1)
    ])
  let box = CGRect(x: margin, y: height - top, width: width - margin * 2, height: top - margin / 2)
  let setter = CTFramesetterCreateWithAttributedString(words)
  CTFrameDraw(
    CTFramesetterCreateFrame(setter, CFRange(), CGPath(rect: box, transform: nil), nil), context)

  let room = CGSize(width: width - margin * 2, height: height - top - margin / 2)
  let scale = min(
    room.width / CGFloat(shot.width), room.height / CGFloat(shot.height))
  let drawn = CGSize(width: CGFloat(shot.width) * scale, height: CGFloat(shot.height) * scale)
  let place = CGRect(
    x: (width - drawn.width) / 2, y: margin / 2 + (room.height - drawn.height), width: drawn.width,
    height: drawn.height)
  let radius = drawn.width * 0.06
  context.addPath(CGPath(roundedRect: place, cornerWidth: radius, cornerHeight: radius, transform: nil))
  context.clip()
  context.interpolationQuality = .high
  context.draw(shot, in: place)
  guard let image = context.makeImage() else { fail("no image") }
  return image
}

/// RGBA bytes of an image at its own size.
func pixels(of image: CGImage) -> [UInt8] {
  var bytes = [UInt8](repeating: 0, count: image.width * image.height * 4)
  bytes.withUnsafeMutableBytes { buffer in
    let context = CGContext(
      data: buffer.baseAddress, width: image.width, height: image.height, bitsPerComponent: 8,
      bytesPerRow: image.width * 4, space: CGColorSpace(name: CGColorSpace.sRGB)!,
      bitmapInfo: CGImageAlphaInfo.noneSkipLast.rawValue)
    context?.draw(image, in: CGRect(x: 0, y: 0, width: image.width, height: image.height))
  }
  return bytes
}

/// The share of pixels that visibly differ, or 1 when the file is missing or another size.
func change(from url: URL, to image: CGImage) -> Double {
  guard let source = CGImageSourceCreateWithURL(url as CFURL, nil),
    let old = CGImageSourceCreateImageAtIndex(source, 0, nil),
    old.width == image.width, old.height == image.height
  else { return 1 }
  let before = pixels(of: old)
  let after = pixels(of: image)
  var changed = 0
  for pixel in stride(from: 0, to: before.count, by: 4) {
    let isSame = (0..<3).allSatisfy { abs(Int(before[pixel + $0]) - Int(after[pixel + $0])) < 12 }
    if !isSame { changed += 1 }
  }
  return Double(changed) / Double(before.count / 4)
}

/// Writes the image unless it only moved a blinking cursor or some glass, as two runs always do.
func save(_ image: CGImage, to url: URL) {
  guard change(from: url, to: image) >= keepsBelow else { return }
  guard
    let file = CGImageDestinationCreateWithURL(url as CFURL, UTType.png.identifier as CFString, 1, nil)
  else { fail("can't write \(url.path)") }
  CGImageDestinationAddImage(file, image, nil)
  guard CGImageDestinationFinalize(file) else { fail("can't write \(url.path)") }
}

let arguments = CommandLine.arguments
guard arguments.count == 5 else { fail("usage: <config> <frames> <captures> <output>") }
let config = try JSONDecoder().decode(
  Config.self, from: Data(contentsOf: URL(filePath: arguments[1])))
let frames = try String(contentsOfFile: arguments[2], encoding: .utf8)
  .split(separator: "\n").map { $0.split(separator: "\t").map(String.init) }
let captures = URL(filePath: arguments[3])
let output = URL(filePath: arguments[4])
let tint = config.background
let background = CGColor(srgbRed: tint[0], green: tint[1], blue: tint[2], alpha: 1)
for size in config.sizes {
  let folder = output.appending(path: size.folder)
  try FileManager.default.createDirectory(at: folder, withIntermediateDirectories: true)
  for frame in frames {
    let capture = load(captures.appending(path: "\(size.folder)/\(frame[0]).png"))
    let shown = screen(in: capture, for: size, name: "\(size.folder)/\(frame[1])")
    let image = compose(shown, headline: frame[2], size: size, background: background)
    save(image, to: folder.appending(path: "\(frame[1]).png"))
  }
}

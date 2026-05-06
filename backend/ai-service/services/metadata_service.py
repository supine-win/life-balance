"""
LifeRecorder AI 服务 - 元数据提取服务

核心功能：
1. 照片 EXIF/IPTC/XMP 元数据提取
2. 视频元数据提取（ffprobe）
3. GPS 逆地理编码
4. 文件名日期模式匹配
5. 生成结构化建议卡片
"""
import os
import re
from datetime import datetime
from typing import Any, Dict, List, Optional, Tuple

from loguru import logger

from core.config import settings
from models.schemas import (
    EXIFData, GPSCoordinate, MetadataExtractionResult,
    MetadataSuggestion, VideoMetadata
)


class MetadataExtractor:
    """多媒体元数据提取器"""

    def __init__(self):
        self._geocode_cache: Dict[str, str] = {}

    async def extract(self, media_id: str, file_path: str, original_name: str) -> MetadataExtractionResult:
        """
        提取媒体文件元数据并生成建议卡片

        Args:
            media_id: 媒体 ID
            file_path: 文件路径
            original_name: 原始文件名

        Returns:
            MetadataExtractionResult: 提取结果
        """
        result = MetadataExtractionResult(media_id=media_id)

        ext = os.path.splitext(original_name)[1].lower()

        try:
            if ext in ('.jpg', '.jpeg', '.png', '.gif', '.webp', '.bmp', '.tiff', '.tif', '.heic', '.heif'):
                await self._extract_image_metadata(result, file_path, original_name)
            elif ext in ('.mp4', '.mov', '.avi', '.mkv', '.wmv', '.flv', '.webm', '.m4v', '.3gp'):
                await self._extract_video_metadata(result, file_path, original_name)
            elif ext in ('.mp3', '.wav', '.ogg', '.flac', '.aac', '.m4a', '.wma'):
                await self._extract_audio_metadata(result, file_path, original_name)

        except Exception as e:
            logger.error(f"元数据提取失败: {e}")
            result.extract_error = str(e)

        # 尝试从文件名提取日期
        if result.shot_time is None:
            date_time, pattern = self._parse_date_from_filename(original_name)
            if date_time:
                result.shot_time = date_time
                result.shot_time_src = "filename"
                result.filename_pattern = pattern

        # GPS 逆地理编码
        if result.location_name == "" and hasattr(result, '_gps'):
            gps = result._gps
            if gps:
                location = await self._reverse_geocode(gps.latitude, gps.longitude)
                if location:
                    result.location_name = location
                    result.location_src = "geocode"

        # 生成建议
        result.suggestions = self._generate_suggestions(result)

        if result.extract_error == "":
            result.extract_status = "completed"
        else:
            result.extract_status = "partial" if result.shot_time or result.camera_model else "failed"

        return result

    async def _extract_image_metadata(self, result: MetadataExtractionResult, file_path: str, original_name: str):
        """提取图片 EXIF 元数据"""
        try:
            import exifread
        except ImportError:
            logger.warning("exifread 未安装，跳过 EXIF 提取")
            return

        with open(file_path, 'rb') as f:
            tags = exifread.process_file(f, details=True)

        if not tags:
            return

        # 保存原始 EXIF
        raw_exif = {}
        for tag, value in tags.items():
            raw_exif[tag] = str(value)
        result.raw_exif = raw_exif

        # 提取拍摄时间
        for dt_tag in ['EXIF DateTimeOriginal', 'Image DateTime', 'EXIF DateTimeDigitized']:
            if dt_tag in tags:
                dt_str = str(tags[dt_tag])
                try:
                    dt = datetime.strptime(dt_str, '%Y:%m:%d %H:%M:%S')
                    result.shot_time = dt
                    result.shot_time_src = "exif"
                    break
                except ValueError:
                    pass

        # 提取 GPS
        gps = self._extract_gps(tags)
        if gps:
            result._gps = gps

        # 提取相机信息
        if 'Image Make' in tags:
            result.camera_make = str(tags['Image Make'])
        if 'Image Model' in tags:
            result.camera_model = str(tags['Image Model'])
        if 'EXIF LensModel' in tags:
            result.lens_model = str(tags['EXIF LensModel'])
        if 'EXIF FocalLength' in tags:
            try:
                result.focal_length = float(str(tags['EXIF FocalLength']).split('/')[0])
            except (ValueError, IndexError):
                pass
        if 'EXIF FNumber' in tags:
            try:
                parts = str(tags['EXIF FNumber']).split('/')
                if len(parts) == 2:
                    result.aperture = float(parts[0]) / float(parts[1])
                else:
                    result.aperture = float(parts[0])
            except (ValueError, IndexError):
                pass
        if 'EXIF ISOSpeedRatings' in tags:
            try:
                result.iso = int(str(tags['EXIF ISOSpeedRatings']))
            except ValueError:
                pass
        if 'EXIF ExposureTime' in tags:
            result.shutter_speed = str(tags['EXIF ExposureTime'])

        # 提取图片尺寸
        try:
            from PIL import Image
            with Image.open(file_path) as img:
                pass  # width/height 可以在这里获取
        except Exception:
            pass

    def _extract_gps(self, tags: dict) -> Optional[GPSCoordinate]:
        """从 EXIF tags 提取 GPS 坐标"""
        def _convert_to_degrees(value):
            """将 GPS 度分秒转换为十进制度数"""
            d = float(value.values[0].num) / float(value.values[0].den)
            m = float(value.values[1].num) / float(value.values[1].den)
            s = float(value.values[2].num) / float(value.values[2].den)
            return d + m / 60.0 + s / 3600.0

        lat = lon = None
        if 'GPS GPSLatitude' in tags and 'GPS GPSLatitudeRef' in tags:
            lat = _convert_to_degrees(tags['GPS GPSLatitude'])
            if str(tags['GPS GPSLatitudeRef']) == 'S':
                lat = -lat

        if 'GPS GPSLongitude' in tags and 'GPS GPSLongitudeRef' in tags:
            lon = _convert_to_degrees(tags['GPS GPSLongitude'])
            if str(tags['GPS GPSLongitudeRef']) == 'W':
                lon = -lon

        if lat and lon:
            alt = None
            if 'GPS GPSAltitude' in tags:
                try:
                    alt = float(str(tags['GPS GPSAltitude']).split('/')[0])
                except (ValueError, IndexError):
                    pass
            return GPSCoordinate(latitude=lat, longitude=lon, altitude=alt)

        return None

    async def _extract_video_metadata(self, result: MetadataExtractionResult, file_path: str, original_name: str):
        """提取视频元数据"""
        try:
            import ffmpeg
        except ImportError:
            logger.warning("ffmpeg-python 未安装，跳过视频元数据提取")
            return

        try:
            probe = ffmpeg.probe(file_path, cmd=settings.FFPROBE_PATH)
        except Exception as e:
            logger.error(f"ffprobe 执行失败: {e}")
            return

        # 提取格式信息
        format_info = probe.get('format', {})
        if 'tags' in format_info:
            tags = format_info['tags']
            if 'creation_time' in tags:
                try:
                    result.shot_time = datetime.fromisoformat(tags['creation_time'].replace('Z', '+00:00'))
                    result.shot_time_src = "video"
                except ValueError:
                    pass

        # 提取流信息
        for stream in probe.get('streams', []):
            if stream['codec_type'] == 'video':
                result.video_codec = stream.get('codec_name', '')
                result.frame_rate = float(stream.get('r_frame_rate', '0').split('/')[0])
            elif stream['codec_type'] == 'audio':
                result.audio_codec = stream.get('codec_name', '')

    async def _extract_audio_metadata(self, result: MetadataExtractionResult, file_path: str, original_name: str):
        """提取音频元数据"""
        # 基础实现：使用 ffprobe
        try:
            import ffmpeg
            probe = ffmpeg.probe(file_path, cmd=settings.FFPROBE_PATH)
            format_info = probe.get('format', {})
            duration = float(format_info.get('duration', 0))
            if duration > 0:
                result.frame_rate = duration  # 借用字段存时长
        except Exception:
            pass

    def _parse_date_from_filename(self, filename: str) -> Tuple[Optional[datetime], str]:
        """
        从文件名解析日期

        支持格式:
        - IMG_20260501_143200.jpg
        - 2024-01-15_hiking.png
        - Screenshot_2024-03-20-15-30-00.png
        - VID_20240501_120000.mp4
        """
        patterns = [
            (r'(?i)(?:IMG|DSC|VID|Screenshot|Photo|photo)\D*?(\d{4})(\d{2})(\d{2})[_\-.]?(\d{2})?(\d{2})?(\d{2})?',
             '%Y%m%d %H%M%S', 'camera_standard'),
            (r'(?i)(\d{4})-(\d{2})-(\d{2})[_\-.]?(\d{2})?(\d{2})?(\d{2})?',
             '%Y-%m-%d %H%M%S', 'iso_date'),
            (r'(?i)(\d{4})_(\d{2})_(\d{2})[_\-.]?(\d{2})?(\d{2})?(\d{2})?',
             '%Y_%m_%d %H%M%S', 'underscore_date'),
        ]

        for regex, date_format, pattern_name in patterns:
            match = re.search(regex, filename)
            if match:
                groups = match.groups()
                date_str = f"{groups[0]}{groups[1]}{groups[2]}"
                if groups[3]:
                    date_str += f" {groups[3]}{groups[4] or '00'}{groups[5] or '00'}"
                else:
                    date_str += " 000000"

                try:
                    dt = datetime.strptime(date_str, date_format)
                    # 验证日期合理性
                    if 2000 <= dt.year <= 2099:
                        return dt, pattern_name
                except ValueError:
                    continue

        return None, ""

    async def _reverse_geocode(self, lat: float, lng: float) -> str:
        """
        GPS 逆地理编码

        优先级:
        1. 高德地图 API
        2. 百度地图 API
        3. Nominatim (OpenStreetMap)
        """
        # 检查缓存
        cache_key = f"{lat:.4f},{lng:.4f}"
        if cache_key in self._geocode_cache:
            return self._geocode_cache[cache_key]

        location = ""

        if settings.GEOCODING_API == "amap" and settings.AMAP_API_KEY:
            location = await self._amap_reverse_geocode(lat, lng)
        elif settings.GEOCODING_API == "baidu" and settings.BAIDU_API_KEY:
            location = await self._baidu_reverse_geocode(lat, lng)

        if not location:
            location = await self._nominatim_reverse_geocode(lat, lng)

        # 缓存结果
        if location:
            self._geocode_cache[cache_key] = location

        return location

    async def _amap_reverse_geocode(self, lat: float, lng: float) -> str:
        """高德地图逆地理编码"""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                resp = await client.get(
                    "https://restapi.amap.com/v3/geocode/regeo",
                    params={
                        "key": settings.AMAP_API_KEY,
                        "location": f"{lng},{lat}",
                        "extensions": "base",
                        "output": "json",
                    },
                    timeout=5,
                )
                data = resp.json()
                if data.get("status") == "1":
                    return data["regeocode"]["formatted_address"]
        except Exception as e:
            logger.warning(f"高德逆地理编码失败: {e}")
        return ""

    async def _baidu_reverse_geocode(self, lat: float, lng: float) -> str:
        """百度地图逆地理编码"""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                resp = await client.get(
                    "https://api.map.baidu.com/reverse_geocoding/v3",
                    params={
                        "ak": settings.BAIDU_API_KEY,
                        "output": "json",
                        "coordtype": "wgs84ll",
                        "location": f"{lat},{lng}",
                    },
                    timeout=5,
                )
                data = resp.json()
                if data.get("status") == 0:
                    return data["result"]["formatted_address"]
        except Exception as e:
            logger.warning(f"百度逆地理编码失败: {e}")
        return ""

    async def _nominatim_reverse_geocode(self, lat: float, lng: float) -> str:
        """Nominatim (OpenStreetMap) 逆地理编码"""
        import httpx
        try:
            async with httpx.AsyncClient() as client:
                resp = await client.get(
                    "https://nominatim.openstreetmap.org/reverse",
                    params={
                        "lat": lat,
                        "lon": lng,
                        "format": "json",
                        "accept-language": "zh",
                        "zoom": 16,
                    },
                    headers={"User-Agent": "LifeRecorder/1.0"},
                    timeout=10,
                )
                data = resp.json()
                if "display_name" in data:
                    return data["display_name"]
        except Exception as e:
            logger.warning(f"Nominatim 逆地理编码失败: {e}")
        return ""

    def _generate_suggestions(self, result: MetadataExtractionResult) -> MetadataSuggestion:
        """生成结构化建议卡片"""
        sug = MetadataSuggestion()
        confidence = {}

        # 建议时间
        if result.shot_time:
            sug.event_time = result.shot_time.isoformat()
            confidence["time"] = 0.95 if result.shot_time_src == "exif" else 0.7

        # 建议地点
        if result.location_name:
            sug.location = result.location_name
            confidence["location"] = 0.88

        # 建议标签
        tags = []
        if result.location_name:
            # 简单标签提取（生产环境可用 NLP）
            parts = result.location_name.replace("市", "").replace("区", "").replace("省", "")
            tags.extend([p for p in parts.split() if len(p) >= 2][:3])
        if result.camera_model:
            tags.append("摄影")
        sug.tags = tags
        confidence["tags"] = 0.65

        # 建议分类
        if result.camera_model:
            sug.category = "摄影"

        # 提示文字
        hints = []
        if result.camera_make and result.camera_model:
            hints.append(f"使用 {result.camera_make} {result.camera_model} 拍摄")
        elif result.camera_model:
            hints.append(f"使用 {result.camera_model} 拍摄")
        if result.location_name:
            hints.append(f"在{result.location_name}")
        sug.hint_text = "".join(hints)

        # 设备信息文字
        device_parts = []
        if result.camera_model:
            device_parts.append(result.camera_model)
        if result.focal_length:
            device_parts.append(f"{result.focal_length:.0f}mm")
        if result.aperture:
            device_parts.append(f"ƒ/{result.aperture:.1f}")
        if result.iso:
            device_parts.append(f"ISO {result.iso}")
        sug.device_text = " · ".join(device_parts)

        sug.confidence = confidence
        return sug

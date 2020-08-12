package constants

// поля помечаются статусом unknown, если их не удалось определить во время парсинга
const UNKNOWN = "unknown"
const REQUEST_URI_DELIM = "/"
const LOG_DELIM = "|"

// данное значение приходит в строке лога, если данные отсутствуют изначально
const EMPTY_VALUE = "-"

// возможно надо найти другой способ для определения "правильности" request uri
const LEN_STREAM_PARTS = 6

const DEFAULT_LOG_FILE = "./tmp/log"
const DEFAULT_MAXMIND_DATABASE = "/usr/share/GeoIP/GeoLite2-City.mmdb"

// константы частей лога, всего из 22 в качестве значений указываются ИНДЕКСЫ 0..21
const FULL_LEN_OF_PARTS = 22

const POS_REMOTE_ADDR = 2
const POS_HOST = 5
const POS_URI = 6
const POS_BYTES_SENT = 21
const POS_CONNECTIONS = 20
const POS_CONNECTION_REQUESTS = 20
const POST_HTTP_SENT_X = 18
const POS_HTTP_USER_AGENT = 17
const POS_HTTP_XFORWARD = 16
const POS_HTTP_VIA = 15
const POS_HTTP_REFERER = 14

// сообщения

const ADDRESS_IS_NOT_DEFINED = "Кажется вы забыли указать адрес для подключения к syslog, пожалуйста, воспользуйтесь аргументом --help"
const NOT_RECOGNIZE_LOGS = "Не удалось распознать строку лога"
const INVALID_PARTS_LENGHT = "Не валидная длина логов Nginx, пожалуйста, ознакомтесь с форматом логов"

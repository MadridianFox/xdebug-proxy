<?php
$a = 12;
$b = 54;

$c = $a + $b;
for ($i = 0; $i < 10; $i++) {
    $c += $i * $a - $b;
}
echo $c;